package github

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/shurcooL/githubv4"
)

func dataSourceGithubTeam() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			// Input
			TEAM_SLUG: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "",
			},
			// Computed
			TEAM_CHILD_TEAMS: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						TEAM_ID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						TEAM_SLUG: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			TEAM_MEMBERS: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						USER_ID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						USER_IS_SITE_ADMIN: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						USER_LOGIN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						USER_NAME: {
							Type:     schema.TypeString,
							Computed: true,
						},
						USER_ROLE: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			TEAM_PARENT_TEAM: {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						TEAM_ID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						TEAM_SLUG: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			TEAM_DESCRIPTION: {
				Type:     schema.TypeString,
				Computed: true,
			},
			TEAM_NAME: {
				Type:     schema.TypeString,
				Computed: true,
			},
			TEAM_PRIVACY: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Read: dataSourceGithubTeamRead,
	}
}

func dataSourceGithubTeamRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Organization struct {
			Team Team `graphql:"team(slug: $slug)"`
		} `graphql:"organization(login: $login)"`
	}
	variables := map[string]interface{}{
		"login":           githubv4.String(meta.(*Organization).Name),
		"slug":            githubv4.String(d.Get(TEAM_SLUG).(string)),
		"childTeamFirst":  githubv4.Int(10),
		"childTeamCursor": (*githubv4.String)(nil),
		"immediateOnly":   githubv4.Boolean(true),
		"membersFirst":    githubv4.Int(10),
		"membersCursor":   (*githubv4.String)(nil),
	}
	var allMembers []struct {
		Node User
		Role githubv4.TeamMemberRole
	}
	var allChildren []struct {
		ID   githubv4.ID
		Slug githubv4.String
	}
	ctx := context.Background()
	client := meta.(*Organization).Client
	for {
		err := client.Query(ctx, &query, variables)
		if err != nil {
			return err
		}

		allMembers = append(allMembers, query.Organization.Team.Members.Edges...)
		allChildren = append(allChildren, query.Organization.Team.ChildTeams.Nodes...)

		if !query.Organization.Team.Members.PageInfo.HasNextPage && !query.Organization.Team.ChildTeams.PageInfo.HasNextPage {
			break
		}

		childTeamCursor := query.Organization.Team.ChildTeams.PageInfo.EndCursor
		if childTeamCursor != "" {
			variables["childTeamCursor"] = childTeamCursor
		} else {
			variables["childTeamCursor"] = (*githubv4.String)(nil)
		}

		membersCursor := query.Organization.Team.Members.PageInfo.EndCursor
		if membersCursor != "" {
			variables["membersCursor"] = membersCursor
		} else {
			variables["membersCursor"] = (*githubv4.String)(nil)
		}
	}

	var childTeams []map[string]interface{}
	for _, t := range allChildren {
		team := make(map[string]interface{})
		team[TEAM_ID] = fmt.Sprintf("%s", t.ID)
		team[TEAM_SLUG] = string(t.Slug)
		childTeams = append(childTeams, team)
	}
	err := d.Set(TEAM_CHILD_TEAMS, childTeams)
	if err != nil {
		return err
	}

	var members []map[string]interface{}
	for _, m := range allMembers {
		member := make(map[string]interface{})
		member[USER_ID] = fmt.Sprintf("%s", m.Node.ID)
		member[USER_IS_SITE_ADMIN] = bool(m.Node.IsSiteAdmin)
		member[USER_LOGIN] = string(m.Node.Login)
		member[USER_NAME] = string(m.Node.Name)
		member[USER_ROLE] = string(m.Role)
		members = append(members, member)
	}
	err = d.Set(TEAM_MEMBERS, members)
	if err != nil {
		return err
	}

	parentTeam := make(map[string]interface{})
	parentTeam[TEAM_ID] = fmt.Sprintf("%s", query.Organization.Team.ParentTeam.ID)
	parentTeam[TEAM_SLUG] = string(query.Organization.Team.ParentTeam.Slug)
	err = d.Set(TEAM_PARENT_TEAM, parentTeam)
	if err != nil {
		return err
	}

	err = d.Set(TEAM_DESCRIPTION, query.Organization.Team.Description)
	if err != nil {

	}

	err = d.Set(TEAM_NAME, query.Organization.Team.Name)
	if err != nil {

	}

	err = d.Set(TEAM_PRIVACY, query.Organization.Team.Privacy)
	if err != nil {

	}

	d.SetId(fmt.Sprintf("%s", query.Organization.Team.ID))

	return nil
}
