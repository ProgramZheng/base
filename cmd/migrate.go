/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/programzheng/base/pkg/model"
	"github.com/programzheng/base/pkg/model/admin"
	"github.com/programzheng/base/pkg/model/auth"
	"github.com/programzheng/base/pkg/model/file"
	"github.com/programzheng/base/pkg/model/post"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !model.HasTable(&admin.Admin{}) {
			model.CreateTable(&admin.Admin{})
		}
		if !model.HasTable(&admin.AdminProfile{}) {
			model.CreateTable(&admin.AdminProfile{})
		}
		if !model.HasTable(&auth.AdminLogin{}) {
			model.CreateTable(&auth.AdminLogin{})
		}
		if !model.HasTable(&file.File{}) {
			model.CreateTable(&file.File{})
		}
		if !model.HasTable(&post.Post{}) {
			model.CreateTable(&post.Post{})
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// migrateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// migrateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
