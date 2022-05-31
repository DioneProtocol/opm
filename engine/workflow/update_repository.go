package workflow

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/version"
	"github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/yaml.v3"

	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/types"
	"github.com/ava-labs/apm/url"
	"github.com/ava-labs/apm/util"
)

var (
	subnetDir = "subnets"
	vmDir     = "vms"

	subnetKey = "subnet"
	vmKey     = "vm"

	_ Workflow = &UpdateRepository{}
)

type UpdateRepositoryConfig struct {
	Executor       Executor
	RepoName       string
	RepositoryPath string

	AliasBytes []byte

	PreviousCommit plumbing.Hash
	LatestCommit   plumbing.Hash

	SourceInfo   storage.SourceInfo
	Repository   storage.Repository
	Registry     storage.Storage[storage.RepoList]
	SourceList   storage.Storage[storage.SourceInfo]
	InstalledVMs storage.Storage[version.Semantic]

	DB database.Database

	TmpPath    string
	PluginPath string
	HttpClient url.Client
}

func NewUpdateRepository(config UpdateRepositoryConfig) *UpdateRepository {
	return &UpdateRepository{
		executor:           config.Executor,
		repoName:           config.RepoName,
		repositoryPath:     config.RepositoryPath,
		aliasBytes:         config.AliasBytes,
		previousCommit:     config.PreviousCommit,
		latestCommit:       config.LatestCommit,
		repository:         config.Repository,
		registry:           config.Registry,
		repositoryMetadata: config.SourceInfo,
		installedVMs:       config.InstalledVMs,
		sourceList:         config.SourceList,
		db:                 config.DB,
		tmpPath:            config.TmpPath,
		pluginPath:         config.PluginPath,
		httpClient:         config.HttpClient,
	}
}

type UpdateRepository struct {
	executor       Executor
	repoName       string
	repositoryPath string

	aliasBytes []byte

	previousCommit plumbing.Hash
	latestCommit   plumbing.Hash

	repository storage.Repository
	registry   storage.Storage[storage.RepoList]

	repositoryMetadata storage.SourceInfo

	installedVMs storage.Storage[version.Semantic]
	sourceList   storage.Storage[storage.SourceInfo]
	db           database.Database

	tmpPath    string
	pluginPath string

	httpClient url.Client
}

func (u *UpdateRepository) updateVMs() error {
	updated := false

	itr := u.installedVMs.Iterator()
	defer itr.Release()

	for itr.Next() {
		fullVMName := string(itr.Key())
		installedVersion, err := itr.Value()
		if err != nil {
			return err
		}

		repoAlias, vmName := util.ParseQualifiedName(fullVMName)
		organization, repo := util.ParseAlias(repoAlias)

		// weird hack for generics to make the go compiler happy.
		var definition storage.Definition[types.VM]

		vmStorage := u.repository.VMs()
		definition, err = vmStorage.Get([]byte(vmName))
		if err != nil {
			return err
		}

		updatedVM := definition.Definition

		if installedVersion.Compare(updatedVM.Version) < 0 {
			fmt.Printf(
				"Detected an update for %s from v%v.%v.%v to v%v.%v.%v.\n",
				fullVMName,
				installedVersion.Major,
				installedVersion.Minor,
				installedVersion.Patch,
				updatedVM.Version.Major,
				updatedVM.Version.Minor,
				updatedVM.Version.Patch,
			)
			installWorkflow := NewInstallWorkflow(InstallWorkflowConfig{
				Name:         fullVMName,
				Plugin:       vmName,
				Organization: organization,
				Repo:         repo,
				TmpPath:      u.tmpPath,
				PluginPath:   u.pluginPath,
				InstalledVMs: u.installedVMs,
				VMStorage:    u.repository.VMs(),
				HttpClient:   u.httpClient,
			})

			fmt.Printf(
				"Rebuilding binaries for %s v%v.%v.%v.\n",
				fullVMName,
				updatedVM.Version.Major,
				updatedVM.Version.Minor,
				updatedVM.Version.Patch,
			)
			if err := u.executor.Execute(installWorkflow); err != nil {
				return err
			}

			updated = true
		}
	}

	if !updated {
		fmt.Printf("No changes detected.")
		return nil
	}

	return nil
}

func (u *UpdateRepository) Execute() error {
	if err := u.updateDefinitions(); err != nil {
		fmt.Printf("Unexpected error while updating definitions. %s", err)
		return err
	}

	if err := u.updateVMs(); err != nil {
		fmt.Printf("Unexpected error while updating vms. %s", err)
		return err
	}

	// checkpoint progress
	updatedMetadata := storage.SourceInfo{
		Alias:  u.repositoryMetadata.Alias,
		URL:    u.repositoryMetadata.URL,
		Commit: u.latestCommit,
	}
	if err := u.sourceList.Put(u.aliasBytes, updatedMetadata); err != nil {
		return err
	}

	fmt.Printf("Finished update.\n")

	return nil
}

func (u *UpdateRepository) updateDefinitions() error {
	vmsPath := filepath.Join(u.repositoryPath, vmDir)

	if err := loadFromYAML[types.VM](vmKey, vmsPath, u.aliasBytes, u.latestCommit, u.registry, u.repository.VMs()); err != nil {
		return err
	}

	subnetsPath := filepath.Join(u.repositoryPath, subnetDir)
	if err := loadFromYAML[types.Subnet](subnetKey, subnetsPath, u.aliasBytes, u.latestCommit, u.registry, u.repository.Subnets()); err != nil {
		return err
	}

	// Now we need to delete anything that wasn't updated in the latest commit
	if err := deleteStalePlugins[types.VM](u.repository.VMs(), u.latestCommit); err != nil {
		return err
	}
	if err := deleteStalePlugins[types.Subnet](u.repository.Subnets(), u.latestCommit); err != nil {
		return err
	}

	if u.previousCommit == plumbing.ZeroHash {
		fmt.Printf("Finished initializing definitions for%s@%s.\n", u.repoName, u.latestCommit)
	} else {
		fmt.Printf("Finished updating definitions from %s to %s@%s.\n", u.previousCommit, u.repoName, u.latestCommit)
	}

	return nil
}

func loadFromYAML[T types.Definition](
	key string,
	path string,
	repositoryAlias []byte,
	commit plumbing.Hash,
	globalDB storage.Storage[storage.RepoList],
	repoDB storage.Storage[storage.Definition[T]],
) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	//globalBatch := globalDB.NewBatch()
	//repoBatch := repoDB.NewBatch()

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		nameWithExtension := file.Name()
		// Strip any extension from the file. This is to support windows .exe
		// files.
		name := nameWithExtension[:len(nameWithExtension)-len(filepath.Ext(nameWithExtension))]

		// Skip hidden files.
		if len(name) == 0 {
			continue
		}

		fileBytes, err := os.ReadFile(filepath.Join(path, file.Name()))
		if err != nil {
			return err
		}
		data := make(map[string]T)

		if err := yaml.Unmarshal(fileBytes, data); err != nil {
			return err
		}
		definition := storage.Definition[T]{
			Definition: data[key],
			Commit:     commit,
		}

		alias := data[key].Alias()
		aliasBytes := []byte(alias)

		registry := storage.RepoList{}
		registry, err = globalDB.Get(aliasBytes)
		if err == database.ErrNotFound {
			registry = storage.RepoList{ //TODO check if this can be removed
				Repositories: []string{},
			}
		} else if err != nil {
			return err
		}

		registry.Repositories = append(registry.Repositories, string(repositoryAlias))

		//updatedRegistryBytes, err := yaml.Marshal(registry)
		//if err != nil {
		//	return err
		//}
		//updatedRecordBytes, err := yaml.Marshal(definition)
		//if err != nil {
		//	return err
		//}

		if err := globalDB.Put(aliasBytes, registry); err != nil {
			return err
		}
		if err := repoDB.Put(aliasBytes, definition); err != nil {
			return err
		}

		fmt.Printf("Updated plugin definition in registry for %s:%s@%s.\n", repositoryAlias, alias, commit)
	}

	//if err := globalBatch.Write(); err != nil {
	//	return err
	//}
	//if err := repoBatch.Write(); err != nil {
	//	return err
	//}
	//
	return nil
}

func deleteStalePlugins[T types.Definition](db storage.Storage[storage.Definition[T]], latestCommit plumbing.Hash) error {
	itr := db.Iterator()
	//TODO batching

	for itr.Next() {
		definition, err := itr.Value()
		if err != nil {
			return err
		}

		if definition.Commit != latestCommit {
			fmt.Printf("Deleting a stale plugin: %s@%s as of %s.\n", definition.Definition.Alias(), definition.Commit, latestCommit)
			if err := db.Delete(itr.Key()); err != nil {
				return err
			}
		}
	}

	return nil
}
