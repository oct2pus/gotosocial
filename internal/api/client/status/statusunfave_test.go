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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/status"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusUnfaveTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusUnfaveTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *StatusUnfaveTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestPostgres()
	suite.storage = testrig.NewTestStorage()
	suite.log = testrig.NewTestLog()
	suite.federator = testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db), suite.storage)
	suite.processor = testrig.NewTestProcessor(suite.db, suite.storage, suite.federator)
	suite.statusModule = status.New(suite.config, suite.processor, suite.log).(*status.Module)
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")
}

func (suite *StatusUnfaveTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

// unfave a status
func (suite *StatusUnfaveTestSuite) TestPostUnfave() {

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// this is the status we wanna unfave: in the testrig it's already faved by this account
	targetStatus := suite.testStatuses["admin_account_status_1"]

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(status.UnfavouritePath, ":id", targetStatus.ID, 1)), nil) // the endpoint we're hitting

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   status.IDKey,
			Value: targetStatus.ID,
		},
	}

	suite.statusModule.StatusUnfavePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &model.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), targetStatus.ContentWarning, statusReply.SpoilerText)
	assert.Equal(suite.T(), targetStatus.Content, statusReply.Content)
	assert.False(suite.T(), statusReply.Sensitive)
	assert.Equal(suite.T(), model.VisibilityPublic, statusReply.Visibility)
	assert.False(suite.T(), statusReply.Favourited)
	assert.Equal(suite.T(), 0, statusReply.FavouritesCount)
}

// try to unfave a status that's already not faved
func (suite *StatusUnfaveTestSuite) TestPostAlreadyNotFaved() {

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// this is the status we wanna unfave: in the testrig it's not faved by this account
	targetStatus := suite.testStatuses["admin_account_status_2"]

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(status.UnfavouritePath, ":id", targetStatus.ID, 1)), nil) // the endpoint we're hitting

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   status.IDKey,
			Value: targetStatus.ID,
		},
	}

	suite.statusModule.StatusUnfavePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &model.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), targetStatus.ContentWarning, statusReply.SpoilerText)
	assert.Equal(suite.T(), targetStatus.Content, statusReply.Content)
	assert.True(suite.T(), statusReply.Sensitive)
	assert.Equal(suite.T(), model.VisibilityPublic, statusReply.Visibility)
	assert.False(suite.T(), statusReply.Favourited)
	assert.Equal(suite.T(), 0, statusReply.FavouritesCount)
}

func TestStatusUnfaveTestSuite(t *testing.T) {
	suite.Run(t, new(StatusUnfaveTestSuite))
}
