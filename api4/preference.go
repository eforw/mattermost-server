// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-server/v6/audit"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
)

func (api *API) InitPreference() {
	api.BaseRoutes.Preferences.Handle("", api.ApiSessionRequired(getPreferences)).Methods("GET")
	api.BaseRoutes.Preferences.Handle("", api.ApiSessionRequired(updatePreferences)).Methods("PUT")
	api.BaseRoutes.Preferences.Handle("/delete", api.ApiSessionRequired(deletePreferences)).Methods("POST")
	api.BaseRoutes.Preferences.Handle("/{category:[A-Za-z0-9_]+}", api.ApiSessionRequired(getPreferencesByCategory)).Methods("GET")
	api.BaseRoutes.Preferences.Handle("/{category:[A-Za-z0-9_]+}/name/{preference_name:[A-Za-z0-9_]+}", api.ApiSessionRequired(getPreferenceByCategoryAndName)).Methods("GET")
}

func getPreferences(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToUser(*c.AppContext.Session(), c.Params.UserId) {
		c.SetPermissionError(model.PermissionEditOtherUsers)
		return
	}

	preferences, err := c.App.GetPreferencesForUser(c.Params.UserId)
	if err != nil {
		c.Err = err
		return
	}

	if err := json.NewEncoder(w).Encode(preferences); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func getPreferencesByCategory(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId().RequireCategory()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToUser(*c.AppContext.Session(), c.Params.UserId) {
		c.SetPermissionError(model.PermissionEditOtherUsers)
		return
	}

	preferences, err := c.App.GetPreferenceByCategoryForUser(c.Params.UserId, c.Params.Category)
	if err != nil {
		c.Err = err
		return
	}

	if err := json.NewEncoder(w).Encode(preferences); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func getPreferenceByCategoryAndName(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId().RequireCategory().RequirePreferenceName()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToUser(*c.AppContext.Session(), c.Params.UserId) {
		c.SetPermissionError(model.PermissionEditOtherUsers)
		return
	}

	preferences, err := c.App.GetPreferenceByCategoryAndNameForUser(c.Params.UserId, c.Params.Category, c.Params.PreferenceName)
	if err != nil {
		c.Err = err
		return
	}

	if err := json.NewEncoder(w).Encode(preferences); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func updatePreferences(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	auditRec := c.MakeAuditRecord("updatePreferences", audit.Fail)
	defer c.LogAuditRec(auditRec)

	if !c.App.SessionHasPermissionToUser(*c.AppContext.Session(), c.Params.UserId) {
		c.SetPermissionError(model.PermissionEditOtherUsers)
		return
	}

	preferences, err := model.PreferencesFromJson(r.Body)
	if err != nil {
		c.SetInvalidParam("preferences")
		return
	}

	var sanitizedPreferences model.Preferences

	for _, pref := range preferences {
		if pref.Category == model.PreferenceCategoryFlaggedPost {
			post, err := c.App.GetSinglePost(pref.Name)
			if err != nil {
				c.SetInvalidParam("preference.name")
				return
			}

			if !c.App.SessionHasPermissionToChannel(*c.AppContext.Session(), post.ChannelId, model.PermissionReadChannel) {
				c.SetPermissionError(model.PermissionReadChannel)
				return
			}
		}

		sanitizedPreferences = append(sanitizedPreferences, pref)
	}

	if err := c.App.UpdatePreferences(c.Params.UserId, sanitizedPreferences); err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	ReturnStatusOK(w)
}

func deletePreferences(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	auditRec := c.MakeAuditRecord("deletePreferences", audit.Fail)
	defer c.LogAuditRec(auditRec)

	if !c.App.SessionHasPermissionToUser(*c.AppContext.Session(), c.Params.UserId) {
		c.SetPermissionError(model.PermissionEditOtherUsers)
		return
	}

	preferences, err := model.PreferencesFromJson(r.Body)
	if err != nil {
		c.SetInvalidParam("preferences")
		return
	}

	if err := c.App.DeletePreferences(c.Params.UserId, preferences); err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	ReturnStatusOK(w)
}
