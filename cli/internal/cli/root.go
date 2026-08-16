package cli

import (
	"Gel/internal/app"
	"Gel/internal/domain"
	"Gel/internal/storage"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type repository struct {
	workingDir  domain.AbsolutePath
	workspace   *domain.Workspace
	objectStore *storage.ObjectStore
	configStore *storage.ConfigStore
	indexStore  *storage.IndexStore
}

type repositoryProvider struct {
	repository *repository
	err        error
	loaded     bool

	getwd func() (string, error)
}

func newRepositoryProvider() *repositoryProvider {
	return &repositoryProvider{
		getwd: os.Getwd,
	}
}

func (p *repositoryProvider) Load() (*repository, error) {
	if p.loaded {
		return p.repository, p.err
	}

	p.loaded = true

	cwd, err := p.getwd()
	if err != nil {
		p.err = fmt.Errorf("get working directory: %w", err)
		return nil, p.err
	}

	workingDir, err := domain.NewAbsolutePath(cwd)
	if err != nil {
		p.err = fmt.Errorf("create absolute path: %w", err)
		return nil, p.err
	}

	workspace, err := app.DiscoverWorkspace(cwd)
	if err != nil {
		p.err = fmt.Errorf("discover workspace: %w", err)
		return nil, p.err
	}

	p.repository = &repository{
		workingDir:  workingDir,
		workspace:   workspace,
		objectStore: storage.NewObjectStore(workspace.ObjectsDir()),
		configStore: storage.NewConfigStore(workspace.ConfigPath()),
		indexStore:  storage.NewIndexStore(workspace.IndexPath()),
	}
	return p.repository, nil
}

func newRootCommand() *cobra.Command {
	provider := newRepositoryProvider()
	rootCommand := &cobra.Command{
		Use:   "Gel",
		Short: "An Agentic Version Control System",
	}
	rootCommand.AddCommand(
		newInitCommand(),
		newHashObjectCommand(provider),
		newCatFileCommand(provider),
		newConfigCommand(provider),
		newAddCommand(provider),
		newWriteTreeCommand(provider),
		newCommitTreeCommand(provider),
		newReadTreeCommand(provider),
		newLsFilesCommand(provider),
	)
	return rootCommand
}

func Execute() int {
	cmd := newRootCommand()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
		return 1
	}
	return 0
}
