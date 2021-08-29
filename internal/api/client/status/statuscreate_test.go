/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package status_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/status"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusCreateTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusCreateTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *StatusCreateTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestPostgres()
	suite.storage = testrig.NewTestStorage()
	suite.log = testrig.NewTestLog()
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.federator = testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db), suite.storage)
	suite.processor = testrig.NewTestProcessor(suite.db, suite.storage, suite.federator)
	suite.statusModule = status.New(suite.config, suite.processor, suite.log).(*status.Module)
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")
}

func (suite *StatusCreateTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

var statusWithLinksAndTags = `#test alright, should be able to post #links with fragments in them now, let's see........

https://docs.gotosocial.org/en/latest/user_guide/posts/#links

#gotosocial

(tobi remember to pull the docker image challenge)`

// Post a new status with some custom visibility settings
func (suite *StatusCreateTestSuite) TestPostNewStatus() {

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Form = url.Values{
		"status":       {"this is a brand new status! #helloworld"},
		"spoiler_text": {"hello hello"},
		"sensitive":    {"true"},
		"visibility":   {string(model.VisibilityMutualsOnly)},
		"likeable":     {"false"},
		"replyable":    {"false"},
		"federated":    {"false"},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	// check response

	// 1. we should have OK from our call to the function
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &model.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "hello hello", statusReply.SpoilerText)
	assert.Equal(suite.T(), "<p>this is a brand new status! <a href=\"http://localhost:8080/tags/helloworld\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>helloworld</span></a></p>", statusReply.Content)
	assert.True(suite.T(), statusReply.Sensitive)
	assert.Equal(suite.T(), model.VisibilityPrivate, statusReply.Visibility) // even though we set this status to mutuals only, it should serialize to private, because masto has no idea about mutuals_only
	assert.Len(suite.T(), statusReply.Tags, 1)
	assert.Equal(suite.T(), model.Tag{
		Name: "helloworld",
		URL:  "http://localhost:8080/tags/helloworld",
	}, statusReply.Tags[0])

	gtsTag := &gtsmodel.Tag{}
	err = suite.db.GetWhere(context.Background(), []db.Where{{Key: "name", Value: "helloworld"}}, gtsTag)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), statusReply.Account.ID, gtsTag.FirstSeenFromAccountID)
}

func (suite *StatusCreateTestSuite) TestPostAnotherNewStatus() {

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Form = url.Values{
		"status": {statusWithLinksAndTags},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	// check response

	// 1. we should have OK from our call to the function
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &model.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "<p><a href=\"http://localhost:8080/tags/test\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>test</span></a> alright, should be able to post <a href=\"http://localhost:8080/tags/links\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>links</span></a> with fragments in them now, let's see........<br><br><a href=\"https://docs.gotosocial.org/en/latest/user_guide/posts/#links\" rel=\"noopener nofollow noreferrer\" target=\"_blank\">docs.gotosocial.org/en/latest/user_guide/posts/#links</a><br><br><a href=\"http://localhost:8080/tags/gotosocial\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>gotosocial</span></a><br><br>(tobi remember to pull the docker image challenge)</p>", statusReply.Content)
}

func (suite *StatusCreateTestSuite) TestPostNewStatusWithEmoji() {

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Form = url.Values{
		"status": {"here is a rainbow emoji a few times! :rainbow: :rainbow: :rainbow: \n here's an emoji that isn't in the db: :test_emoji: "},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &model.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "", statusReply.SpoilerText)
	assert.Equal(suite.T(), "<p>here is a rainbow emoji a few times! :rainbow: :rainbow: :rainbow:<br>here's an emoji that isn't in the db: :test_emoji:</p>", statusReply.Content)

	assert.Len(suite.T(), statusReply.Emojis, 1)
	mastoEmoji := statusReply.Emojis[0]
	gtsEmoji := testrig.NewTestEmojis()["rainbow"]

	assert.Equal(suite.T(), gtsEmoji.Shortcode, mastoEmoji.Shortcode)
	assert.Equal(suite.T(), gtsEmoji.ImageURL, mastoEmoji.URL)
	assert.Equal(suite.T(), gtsEmoji.ImageStaticURL, mastoEmoji.StaticURL)
}

// Try to reply to a status that doesn't exist
func (suite *StatusCreateTestSuite) TestReplyToNonexistentStatus() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Form = url.Values{
		"status":         {"this is a reply to a status that doesn't exist"},
		"spoiler_text":   {"don't open cuz it won't work"},
		"in_reply_to_id": {"3759e7ef-8ee1-4c0c-86f6-8b70b9ad3d50"},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	// check response

	suite.EqualValues(http.StatusBadRequest, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `{"error":"bad request"}`, string(b))
}

// Post a reply to the status of a local user that allows replies.
func (suite *StatusCreateTestSuite) TestReplyToLocalStatus() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Form = url.Values{
		"status":         {fmt.Sprintf("hello @%s this reply should work!", testrig.NewTestAccounts()["local_account_2"].Username)},
		"in_reply_to_id": {testrig.NewTestStatuses()["local_account_2_status_1"].ID},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &model.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "", statusReply.SpoilerText)
	assert.Equal(suite.T(), fmt.Sprintf("<p>hello <span class=\"h-card\"><a href=\"http://localhost:8080/@%s\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>%s</span></a></span> this reply should work!</p>", testrig.NewTestAccounts()["local_account_2"].Username, testrig.NewTestAccounts()["local_account_2"].Username), statusReply.Content)
	assert.False(suite.T(), statusReply.Sensitive)
	assert.Equal(suite.T(), model.VisibilityPublic, statusReply.Visibility)
	assert.Equal(suite.T(), testrig.NewTestStatuses()["local_account_2_status_1"].ID, statusReply.InReplyToID)
	assert.Equal(suite.T(), testrig.NewTestAccounts()["local_account_2"].ID, statusReply.InReplyToAccountID)
	assert.Len(suite.T(), statusReply.Mentions, 1)
}

// Take a media file which is currently not associated with a status, and attach it to a new status.
func (suite *StatusCreateTestSuite) TestAttachNewMediaSuccess() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	attachment := suite.testAttachments["local_account_1_unattached_1"]

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Form = url.Values{
		"status":    {"here's an image attachment"},
		"media_ids": {attachment.ID},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusResponse := &model.Status{}
	err = json.Unmarshal(b, statusResponse)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "", statusResponse.SpoilerText)
	assert.Equal(suite.T(), "<p>here's an image attachment</p>", statusResponse.Content)
	assert.False(suite.T(), statusResponse.Sensitive)
	assert.Equal(suite.T(), model.VisibilityPublic, statusResponse.Visibility)

	// there should be one media attachment
	assert.Len(suite.T(), statusResponse.MediaAttachments, 1)

	// get the updated media attachment from the database
	gtsAttachment, err := suite.db.GetAttachmentByID(context.Background(), statusResponse.MediaAttachments[0].ID)
	assert.NoError(suite.T(), err)

	// convert it to a masto attachment
	gtsAttachmentAsMasto, err := suite.tc.AttachmentToMasto(context.Background(), gtsAttachment)
	assert.NoError(suite.T(), err)

	// compare it with what we have now
	assert.EqualValues(suite.T(), statusResponse.MediaAttachments[0], gtsAttachmentAsMasto)

	// the status id of the attachment should now be set to the id of the status we just created
	assert.Equal(suite.T(), statusResponse.ID, gtsAttachment.StatusID)
}

func TestStatusCreateTestSuite(t *testing.T) {
	suite.Run(t, new(StatusCreateTestSuite))
}
