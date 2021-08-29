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

package timeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ManagerTestSuite struct {
	TimelineStandardTestSuite
}

func (suite *ManagerTestSuite) SetupSuite() {
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *ManagerTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestPostgres()
	suite.log = testrig.NewTestLog()
	suite.tc = testrig.NewTestTypeConverter(suite.db)

	testrig.StandardDBSetup(suite.db, nil)

	manager := testrig.NewTestTimelineManager(suite.db)
	suite.manager = manager
}

func (suite *ManagerTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *ManagerTestSuite) TestManagerIntegration() {
	testAccount := suite.testAccounts["local_account_1"]

	// should start at 0
	indexedLen := suite.manager.GetIndexedLength(context.Background(), testAccount.ID)
	suite.Equal(0, indexedLen)

	// oldestIndexed should be empty string since there's nothing indexed
	oldestIndexed, err := suite.manager.GetOldestIndexedID(context.Background(), testAccount.ID)
	suite.NoError(err)
	suite.Empty(oldestIndexed)

	// trigger status preparation
	err = suite.manager.PrepareXFromTop(context.Background(), testAccount.ID, 20)
	suite.NoError(err)

	// local_account_1 can see 12 statuses out of the testrig statuses in its home timeline
	indexedLen = suite.manager.GetIndexedLength(context.Background(), testAccount.ID)
	suite.Equal(12, indexedLen)

	// oldest should now be set
	oldestIndexed, err = suite.manager.GetOldestIndexedID(context.Background(), testAccount.ID)
	suite.NoError(err)
	suite.Equal("01F8MH75CBF9JFX4ZAD54N0W0R", oldestIndexed)

	// get hometimeline
	statuses, err := suite.manager.HomeTimeline(context.Background(), testAccount.ID, "", "", "", 20, false)
	suite.NoError(err)
	suite.Len(statuses, 12)

	// now wipe the last status from all timelines, as though it had been deleted by the owner
	err = suite.manager.WipeStatusFromAllTimelines(context.Background(), "01F8MH75CBF9JFX4ZAD54N0W0R")
	suite.NoError(err)

	// timeline should be shorter
	indexedLen = suite.manager.GetIndexedLength(context.Background(), testAccount.ID)
	suite.Equal(11, indexedLen)

	// oldest should now be different
	oldestIndexed, err = suite.manager.GetOldestIndexedID(context.Background(), testAccount.ID)
	suite.NoError(err)
	suite.Equal("01F8MH82FYRXD2RC6108DAJ5HB", oldestIndexed)

	// delete the new oldest status specifically from this timeline, as though local_account_1 had muted or blocked it
	removed, err := suite.manager.Remove(context.Background(), testAccount.ID, "01F8MH82FYRXD2RC6108DAJ5HB")
	suite.NoError(err)
	suite.Equal(2, removed) // 1 status should be removed, but from both indexed and prepared, so 2 removals total

	// timeline should be shorter
	indexedLen = suite.manager.GetIndexedLength(context.Background(), testAccount.ID)
	suite.Equal(10, indexedLen)

	// oldest should now be different
	oldestIndexed, err = suite.manager.GetOldestIndexedID(context.Background(), testAccount.ID)
	suite.NoError(err)
	suite.Equal("01F8MHAAY43M6RJ473VQFCVH37", oldestIndexed)

	// now remove all entries by local_account_2 from the timeline
	err = suite.manager.WipeStatusesFromAccountID(context.Background(), testAccount.ID, suite.testAccounts["local_account_2"].ID)
	suite.NoError(err)

	// timeline should be empty now
	indexedLen = suite.manager.GetIndexedLength(context.Background(), testAccount.ID)
	suite.Equal(5, indexedLen)

	// ingest 1 into the timeline
	status1 := suite.testStatuses["admin_account_status_1"]
	ingested, err := suite.manager.Ingest(context.Background(), status1, testAccount.ID)
	suite.NoError(err)
	suite.True(ingested)

	// ingest and prepare another one into the timeline
	status2 := suite.testStatuses["local_account_2_status_1"]
	ingested, err = suite.manager.IngestAndPrepare(context.Background(), status2, testAccount.ID)
	suite.NoError(err)
	suite.True(ingested)

	// timeline should be longer now
	indexedLen = suite.manager.GetIndexedLength(context.Background(), testAccount.ID)
	suite.Equal(7, indexedLen)

	// try to ingest status 2 again
	ingested, err = suite.manager.IngestAndPrepare(context.Background(), status2, testAccount.ID)
	suite.NoError(err)
	suite.False(ingested) // should be false since it's a duplicate
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}
