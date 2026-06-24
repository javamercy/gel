package cli

import (
	"Gel/internal"
	"Gel/internal/branch"
	"Gel/internal/commit"
	"Gel/internal/core"
	"Gel/internal/diff"
	"Gel/internal/domain"
	"Gel/internal/inspect"
	"Gel/internal/staging"
	"Gel/internal/storage"
	"Gel/internal/tree"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	workspace *domain.Workspace
)
var (
	objectService          *core.ObjectService
	indexService           *core.IndexService
	configService          *core.ConfigService
	refService             *core.RefService
	hashObjectService      *core.HashObjectService
	treeResolver           *core.TreeResolver
	pathResolver           *core.PathResolver
	changeDetector         *core.ChangeDetector
	repositoryPathResolver *core.RepositoryPathResolver
)

var (
	addService         *staging.AddService
	catFileService     *inspect.CatFileService
	lsFilesService     *staging.LsFilesService
	updateIndexService *staging.UpdateIndexService
	writeTreeService   *tree.WriteTreeService
	readTreeService    *tree.ReadTreeService
	lsTreeService      *tree.LsTreeService
	commitTreeService  *commit.CommitTreeService
	symbolicRefService *core.SymbolicRefService
	updateRefService   *core.UpdateRefService
	commitService      *commit.CommitService
	logService         *commit.LogService
	branchService      *branch.BranchService
	restoreService     *inspect.RestoreService
	removeService      *staging.RemoveService
	switchService      *branch.SwitchService
	statusService      *inspect.StatusService
	diffService        *diff.DiffService
	showService        *inspect.ShowService
	commitResolver     *core.CommitResolver
	resetService       *internal.ResetService

	isServicesInitialized bool
)

var commandsWithoutRepository = map[string]bool{
	"init": true,
	"help": true,
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gel",
	Short: "An Agentic Version Control System",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if commandsWithoutRepository[cmd.Name()] {
			return nil
		}
		return initializeServices()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() int {
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(rootCmd.ErrOrStderr(), "error: %v\n", err)
		return 1
	}
	return 0
}

// initializeServices sets up all services lazily when a command needs them
func initializeServices() error {
	if isServicesInitialized {
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	workspace, err = domain.NewWorkspace(cwd)
	if err != nil {
		return err
	}

	baseDir, err := domain.NewAbsolutePath(cwd)
	if err != nil {
		return err
	}

	repositoryPathResolver, err = core.NewRepositoryPathResolver(baseDir, workspace.RepoDir)
	if err != nil {
		return err
	}

	objectStorage := storage.NewObjectStorage(workspace)
	indexStorage := storage.NewIndexStorage(workspace)
	configStorage := storage.NewConfigStorage(workspace)

	objectService = core.NewObjectService(objectStorage)
	indexService = core.NewIndexService(indexStorage)
	configService = core.NewConfigService(configStorage)
	refService = core.NewRefService(workspace)
	hashObjectService = core.NewHashObjectService(objectService)
	pathResolver = core.NewPathResolver(repositoryPathResolver, nil)
	changeDetector = core.NewChangeDetector(objectService, workspace.RepoDir)
	treeResolver = core.NewTreeResolver(
		objectService, indexService, refService, pathResolver, changeDetector, workspace,
	)
	symbolicRefService = core.NewSymbolicRefService(refService)
	updateRefService = core.NewUpdateRefService(refService)

	catFileService = inspect.NewCatFileService(objectService)
	updateIndexService = staging.NewUpdateIndexService(
		indexService, objectService, hashObjectService, changeDetector, workspace,
	)
	addService = staging.NewAddService(indexService, updateIndexService, pathResolver, workspace)
	lsFilesService = staging.NewLsFilesService(indexService, changeDetector, workspace)
	writeTreeService = tree.NewWriteTreeService(indexService, objectService)
	readTreeService = tree.NewReadTreeService(indexService, objectService)
	lsTreeService = tree.NewLsTreeService(objectService)
	commitTreeService = commit.NewCommitTreeService(objectService, configService)
	commitService = commit.NewCommitService(writeTreeService, commitTreeService, refService, objectService)
	logService = commit.NewLogService(refService, objectService)
	branchService = branch.NewBranchService(refService, objectService, workspace)
	switchService = branch.NewSwitchService(
		refService, branchService, objectService, readTreeService, treeResolver, workspace,
	)
	restoreService = inspect.NewRestoreService(
		indexService, objectService, refService, treeResolver, changeDetector, workspace,
	)
	statusService = inspect.NewStatusService(indexService, objectService, branchService, treeResolver)
	diffService = diff.NewDiffService(objectService, treeResolver, diff.NewMyersDiffAlgorithm(), workspace)
	showService = inspect.NewShowService(objectService, refService, diffService)
	commitResolver = core.NewCommitResolver(refService, objectService)
	resetService = internal.NewResetService(
		refService, objectService, readTreeService, treeResolver, commitResolver, workspace,
	)
	removeService = staging.NewRemoveService(indexService, treeResolver, changeDetector, workspace)

	isServicesInitialized = true
	return nil
}
