// Copyright 2020 The PipeCD Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
	oauth2github "golang.org/x/oauth2/github"

	"github.com/pipe-cd/pipecd/pkg/model"
)

// OAuthClient is a oauth client for github.
type OAuthClient struct {
	*github.Client

	project *model.Project

	adminTeam  string
	editorTeam string
	viewerTeam string
}

// NewOAuthClient creates a new oauth client for GitHub.
func NewOAuthClient(ctx context.Context,
	sso *model.ProjectSSOConfig_GitHub,
	rbac *model.ProjectRBACConfig,
	project *model.Project,
	code string,
) (*OAuthClient, error) {
	c := &OAuthClient{
		project:    project,
		adminTeam:  rbac.Admin,
		editorTeam: rbac.Editor,
		viewerTeam: rbac.Viewer,
	}
	cfg := oauth2.Config{
		ClientID:     sso.ClientId,
		ClientSecret: sso.ClientSecret,
		Endpoint:     oauth2github.Endpoint,
	}

	if sso.ProxyUrl != "" {
		proxyURL, err := url.Parse(sso.ProxyUrl)
		if err != nil {
			return nil, err
		}

		t := http.DefaultTransport.(*http.Transport).Clone()
		t.Proxy = http.ProxyURL(proxyURL)
		ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Transport: t})
	}

	if sso.BaseUrl != "" {
		baseURL, err := url.Parse(sso.BaseUrl)
		if err != nil {
			return nil, err
		}
		cfg.Endpoint.TokenURL = fmt.Sprintf("%s://%s%s", baseURL.Scheme, baseURL.Host, "/login/oauth/access_token")

		token, err := cfg.Exchange(ctx, code)
		if err != nil {
			return nil, err
		}

		cli, err := github.NewEnterpriseClient(sso.BaseUrl, sso.UploadUrl, cfg.Client(ctx, token))
		if err != nil {
			return nil, err
		}

		c.Client = cli
		return c, nil
	}

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	c.Client = github.NewClient(cfg.Client(ctx, token))
	return c, nil
}

// GetUser returns a user model.
func (c *OAuthClient) GetUser(ctx context.Context) (*model.User, error) {
	user, _, err := c.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}
	teams, _, err := c.Teams.ListUserTeams(ctx, nil)
	if err != nil {
		return nil, err
	}
	role, err := c.decideRole(user.GetLogin(), teams)
	if err != nil {
		return nil, err
	}

	return &model.User{
		Username:  user.GetLogin(),
		AvatarUrl: user.GetAvatarURL(),
		Role: &model.Role{
			ProjectId:   c.project.Id,
			ProjectRole: role,
		},
	}, nil
}

func (c *OAuthClient) decideRole(user string, teams []*github.Team) (role model.Role_ProjectRole, err error) {
	var found bool

	for _, team := range teams {
		slug := team.GetSlug()
		org := team.Organization.GetLogin()
		if org == "" || slug == "" {
			continue
		}

		t := fmt.Sprintf("%s/%s", org, slug)
		switch t {
		case c.adminTeam:
			role = model.Role_ADMIN
			return
		case c.editorTeam:
			role = model.Role_EDITOR
			found = true
		case c.viewerTeam:
			if role != model.Role_EDITOR {
				role = model.Role_VIEWER
				found = true
			}
		}
	}

	if found {
		return
	}

	// In case the current user does not belong to any registered
	// teams, if AllowStrayAsViewer option is set, assign Viewer role
	// as user's role.
	if c.project.AllowStrayAsViewer {
		role = model.Role_VIEWER
		return
	}

	err = fmt.Errorf("user (%s) not found in any of the %d project teams", user, len(teams))
	return
}
