package usercount

import (
	"github.com/checkmarx/ast-cli/internal/commands/util/printer"
	"github.com/checkmarx/ast-cli/internal/params"
	"github.com/checkmarx/ast-cli/internal/wrappers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type RepositoryView struct {
	Name               string `json:"name"`
	UniqueContributors uint64 `json:"unique_contributors"`
}

const (
	GithubCommand  = "github"
	githubShort    = "The github command presents the unique contributors for the provided GitHub repositories or organizations."
	ReposFlag      = "repos"
	reposFlagUsage = "List of repositories to scan for contributors"
	OrgsFlag       = "orgs"
	orgsFlagUsage  = "List of organizations to scan for contributors"
	githubAPIURL   = "https://api.github.com"
	sinceParam     = "since"
	missingArgs    = "provide at least one repository or organization"
	missingOrg     = "an organization is required for your repositories"
	tooManyOrgs    = "a single organization should be provided for specific repositories"
	botType        = "Bot"
)

var (
	repos, orgs []string
)

func newUserCountGithubCommand(gitHubWrapper wrappers.GitHubWrapper) *cobra.Command {
	userCountCmd := &cobra.Command{
		Use:     GithubCommand,
		Short:   githubShort,
		PreRunE: preRunGithubUserCount,
		RunE:    createRunGitHubUserCountFunc(gitHubWrapper),
	}

	userCountCmd.Flags().StringSliceVar(&repos, ReposFlag, []string{}, reposFlagUsage)
	userCountCmd.Flags().StringSliceVar(&orgs, OrgsFlag, []string{}, orgsFlagUsage)
	userCountCmd.Flags().String(params.GitHubURLFlag, githubAPIURL, params.GitHubURLFlagUsage)

	_ = viper.BindPFlag(params.GitHubURLFlag, userCountCmd.Flags().Lookup(params.GitHubURLFlag))

	return userCountCmd
}

func preRunGithubUserCount(*cobra.Command, []string) error {
	if len(repos) == 0 && len(orgs) == 0 {
		return errors.New(missingArgs)
	}

	if len(repos) > 0 && len(orgs) == 0 {
		return errors.New(missingOrg)
	}

	if len(repos) > 0 && len(orgs) > 1 {
		return errors.New(tooManyOrgs)
	}

	return nil
}

func createRunGitHubUserCountFunc(gitHubWrapper wrappers.GitHubWrapper) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		var totalCommits []wrappers.CommitRoot
		var views []RepositoryView

		_ = viper.BindPFlag(params.SCMTokenFlag, cmd.Flags().Lookup(params.SCMTokenFlag))

		if len(repos) > 0 {
			totalCommits, views, err = collectFromRepos(gitHubWrapper)
		} else {
			totalCommits, views, err = collectFromOrgs(gitHubWrapper)
		}
		if err != nil {
			return err
		}

		uniqueContributors := getUniqueContributors(totalCommits)

		views = append(
			views,
			RepositoryView{
				Name:               TotalContributorsName,
				UniqueContributors: uniqueContributors,
			},
		)

		err = printer.Print(cmd.OutOrStdout(), views, format)

		return err
	}
}

func collectFromRepos(gitHubWrapper wrappers.GitHubWrapper) ([]wrappers.CommitRoot, []RepositoryView, error) {
	var totalCommits []wrappers.CommitRoot
	var views []RepositoryView
	for _, repo := range repos {
		repository, err := gitHubWrapper.GetRepository(orgs[0], repo)
		if err != nil {
			return totalCommits, views, err
		}

		commits, err := gitHubWrapper.GetCommits(repository, map[string]string{sinceParam: ninetyDaysDate})
		if err != nil {
			return totalCommits, views, err
		}

		totalCommits = append(totalCommits, commits...)

		uniqueContributors := getUniqueContributors(commits)

		views = append(
			views,
			RepositoryView{
				Name:               repository.FullName,
				UniqueContributors: uniqueContributors,
			},
		)
	}
	return totalCommits, views, nil
}

func collectFromOrgs(gitHubWrapper wrappers.GitHubWrapper) ([]wrappers.CommitRoot, []RepositoryView, error) {
	var totalCommits []wrappers.CommitRoot
	var views []RepositoryView

	for _, org := range orgs {
		organization, err := gitHubWrapper.GetOrganization(org)
		if err != nil {
			return totalCommits, views, err
		}

		repositories, err := gitHubWrapper.GetRepositories(organization)
		if err != nil {
			return totalCommits, views, err
		}

		for _, repository := range repositories {
			commits, err := gitHubWrapper.GetCommits(repository, map[string]string{sinceParam: ninetyDaysDate})
			if err != nil {
				return totalCommits, views, err
			}

			totalCommits = append(totalCommits, commits...)

			uniqueContributors := getUniqueContributors(commits)

			views = append(
				views,
				RepositoryView{
					Name:               repository.FullName,
					UniqueContributors: uniqueContributors,
				},
			)
		}
	}
	return totalCommits, views, nil
}

func getUniqueContributors(commits []wrappers.CommitRoot) uint64 {
	var contributors = map[string]bool{}
	for _, commit := range commits {
		name := commit.Commit.CommitAuthor.Name
		if !contributors[name] && isNotBot(commit) {
			contributors[name] = true
		}
	}
	return uint64(len(contributors))
}

func isNotBot(commit wrappers.CommitRoot) bool {
	return commit.Author == nil || commit.Author.Type != botType
}
