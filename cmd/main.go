package main

import (
	"time"

	"github.com/spf13/cobra"
)

const MAX_RETRY = 3

var (
	dates []string
	src   string
	dest  string
)

func newFetchYahooNewsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yahoo",
		Short: "Fetch yahoo news rss list",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := fetchYahooNewsRSSList(dest, MAX_RETRY)
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&dest, "dest", "~/Desktop", "dest dir path")

	return cmd
}

func newFetchRSSCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rss",
		Short: "Fetch yahoo news rss file",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := fetchYahooNewsRSS(src, dest, MAX_RETRY)
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&src, "src", "~/Desktop", "src dir path")
	cmd.PersistentFlags().StringVar(&dest, "dest", "~/Desktop", "dest dir path")

	return cmd
}

func newTransformJsonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "json",
		Short: "Transform yahoo news rss file to json",
		RunE: withLoggingE(func(cmd *cobra.Command, args []string) error {
			return eachDate(dates, func(date time.Time) error {
				return transformJSON(src, dest, date)
			})
		}),
	}
	cmd.PersistentFlags().StringVar(&src, "src", "~/Desktop", "src dir path")
	cmd.PersistentFlags().StringVar(&dest, "dest", "~/Desktop", "dest dir path")
	setDatesFlag(cmd.Flags(), &dates, "target date")
	_ = cmd.MarkFlagRequired("date")

	return cmd
}

func newTransformAnalysisCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analysis",
		Short: "Transform yahoo news json file to mecab analysis",
		RunE: withLoggingE(func(cmd *cobra.Command, args []string) error {
			return eachDate(dates, func(date time.Time) error {
				return transformAnalysis(src, dest, date)
			})
		}),
	}
	cmd.PersistentFlags().StringVar(&src, "src", "~/Desktop", "src dir path")
	cmd.PersistentFlags().StringVar(&dest, "dest", "~/Desktop", "src dir path")
	setDatesFlag(cmd.Flags(), &dates, "target date")
	_ = cmd.MarkFlagRequired("date")

	return cmd
}

func newTransformMarkdownCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "markdown",
		Short: "Transform mecab analysis json file to markdown",
		Args:  cobra.NoArgs,
		RunE: withLoggingE(func(cmd *cobra.Command, args []string) error {
			return eachDate(dates, func(date time.Time) error {
				return transformMarkdown(src, dest, date)
			})
		}),
	}
	setDatesFlag(cmd.Flags(), &dates, "date for which the URL list file(s) is generated")
	_ = cmd.MarkFlagRequired("date")
	cmd.Flags().StringVar(&src, "src", "~/Desktop", "src dir path")
	cmd.Flags().StringVar(&dest, "dest", "~/Desktop", "dest dir path")

	return cmd
}

func main() {
	rootCmd := &cobra.Command{Use: "fetch"}
	rootCmd.AddCommand(
		newFetchYahooNewsCommand(),
		newFetchRSSCommand(),
		newTransformJsonCommand(),
		newTransformAnalysisCommand(),
		newTransformMarkdownCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
