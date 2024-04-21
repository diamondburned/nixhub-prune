package main

import (
	"fmt"
	"log"
	"os"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
)

const (
	guildID      = 361910177961738242
	rolesBeneath = "---------- Colors ----------"
	pruneDays    = 30
)

func init() {
	log.SetFlags(0)
}

func main() {
	client := api.NewClient(os.Getenv("TOKEN"))

	roles, err := client.Roles(guildID)
	if err != nil {
		log.Fatalln("cannot get roles:", err)
	}

	discord.SortRolesByPosition(roles)

	var rolesBeneathIDs []discord.RoleID
	var rolesBeneathAlready bool

	for _, role := range roles {
		if role.Name == "@everyone" {
			continue
		}
		if role.Name == rolesBeneath {
			rolesBeneathAlready = true
		}
		if rolesBeneathAlready {
			rolesBeneathIDs = append(rolesBeneathIDs, role.ID)
			log.Println("will include role", role.Name, "with ID", role.ID)
		}
	}

	pruneCount, err := pruntCount(client, guildID, pruneCountOptions{
		Days:         pruneDays,
		IncludeRoles: rolesBeneathIDs,
	})
	if err != nil {
		log.Fatalln("cannot get prune count:", err)
	}

	log.Println("will prune", pruneCount, "members total")
	if prompt("do you want to continue? [y/N] ") != "y" {
		log.Fatalln("aborted")
	}

	if err := prune(client, guildID, pruneOptions{
		Days:         pruneDays,
		IncludeRoles: rolesBeneathIDs,
	}); err != nil {
		log.Fatalln("cannot prune:", err)
	}
}

func prompt(prompt string) string {
	var input string
	fmt.Fprint(os.Stderr, prompt)
	_, err := fmt.Scanln(&input)
	if err != nil {
		log.Fatalln("cannot read input:", err)
	}
	return input
}

type pruneOptions struct {
	ComputePruneCount bool             `json:"compute_prune_count"`
	Days              int              `json:"days"`
	IncludeRoles      []discord.RoleID `json:"include_roles"`
}

func prune(c *api.Client, guildID discord.GuildID, opts pruneOptions) error {
	return c.FastRequest(
		"POST",
		api.EndpointGuilds+guildID.String()+"/prune",
		httputil.WithJSONBody(opts),
	)
}

type pruneCountOptions struct {
	Days         int              `schema:"days"`
	IncludeRoles []discord.RoleID `schema:"include_roles"`
}

func pruntCount(c *api.Client, guildID discord.GuildID, opts pruneCountOptions) (int, error) {
	var resp struct {
		Pruned int `json:"pruned"`
	}
	if err := c.RequestJSON(
		&resp,
		"GET",
		api.EndpointGuilds+guildID.String()+"/prune",
		httputil.WithSchema(c.SchemaEncoder, opts),
	); err != nil {
		return 0, err
	}
	return resp.Pruned, nil
}
