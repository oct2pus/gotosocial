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

package text_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

const text1 = `
This is a text with some links in it. Here's link number one: https://example.org/link/to/something#fragment

Here's link number two: http://test.example.org?q=bahhhhhhhhhhhh

https://another.link.example.org/with/a/pretty/long/path/at/the/end/of/it

really.cool.website <-- this one shouldn't be parsed as a link because it doesn't contain the scheme

https://example.orghttps://google.com <-- this shouldn't work either, but it does?! OK
`

const text2 = `
this is one link: https://example.org

this is the same link again: https://example.org

these should be deduplicated
`

const text3 = `
here's a mailto link: mailto:whatever@test.org
`

const text4 = `
two similar links:

https://example.org

https://example.org/test
`

const text5 = `
what happens when we already have a link within an href?

<a href="https://example.org">https://example.org</a>
`

type LinkTestSuite struct {
	TextStandardTestSuite
}

func (suite *LinkTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testTags = testrig.NewTestTags()
}

func (suite *LinkTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestPostgres()
	suite.log = testrig.NewTestLog()
	suite.formatter = text.NewFormatter(suite.config, suite.db, suite.log)

	testrig.StandardDBSetup(suite.db, nil)
}

func (suite *LinkTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *LinkTestSuite) TestParseSimple() {
	f := suite.formatter.FromPlain(context.Background(), simple, nil, nil)
	assert.Equal(suite.T(), simpleExpected, f)
}

func (suite *LinkTestSuite) TestParseURLsFromText1() {
	urls, err := text.FindLinks(text1)

	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "https://example.org/link/to/something#fragment", urls[0].String())
	assert.Equal(suite.T(), "http://test.example.org?q=bahhhhhhhhhhhh", urls[1].String())
	assert.Equal(suite.T(), "https://another.link.example.org/with/a/pretty/long/path/at/the/end/of/it", urls[2].String())
	assert.Equal(suite.T(), "https://example.orghttps://google.com", urls[3].String())
}

func (suite *LinkTestSuite) TestParseURLsFromText2() {
	urls, err := text.FindLinks(text2)
	assert.NoError(suite.T(), err)

	// assert length 1 because the found links will be deduplicated
	assert.Len(suite.T(), urls, 1)
}

func (suite *LinkTestSuite) TestParseURLsFromText3() {
	urls, err := text.FindLinks(text3)
	assert.NoError(suite.T(), err)

	// assert length 0 because `mailto:` isn't accepted
	assert.Len(suite.T(), urls, 0)
}

func (suite *LinkTestSuite) TestReplaceLinksFromText1() {
	replaced := suite.formatter.ReplaceLinks(context.Background(), text1)
	assert.Equal(suite.T(), `
This is a text with some links in it. Here's link number one: <a href="https://example.org/link/to/something#fragment" rel="noopener">example.org/link/to/something#fragment</a>

Here's link number two: <a href="http://test.example.org?q=bahhhhhhhhhhhh" rel="noopener">test.example.org?q=bahhhhhhhhhhhh</a>

<a href="https://another.link.example.org/with/a/pretty/long/path/at/the/end/of/it" rel="noopener">another.link.example.org/with/a/pretty/long/path/at/the/end/of/it</a>

really.cool.website <-- this one shouldn't be parsed as a link because it doesn't contain the scheme

<a href="https://example.orghttps://google.com" rel="noopener">example.orghttps//google.com</a> <-- this shouldn't work either, but it does?! OK
`, replaced)
}

func (suite *LinkTestSuite) TestReplaceLinksFromText2() {
	replaced := suite.formatter.ReplaceLinks(context.Background(), text2)
	assert.Equal(suite.T(), `
this is one link: <a href="https://example.org" rel="noopener">example.org</a>

this is the same link again: <a href="https://example.org" rel="noopener">example.org</a>

these should be deduplicated
`, replaced)
}

func (suite *LinkTestSuite) TestReplaceLinksFromText3() {
	// we know mailto links won't be replaced with hrefs -- we only accept https and http
	replaced := suite.formatter.ReplaceLinks(context.Background(), text3)
	assert.Equal(suite.T(), `
here's a mailto link: mailto:whatever@test.org
`, replaced)
}

func (suite *LinkTestSuite) TestReplaceLinksFromText4() {
	replaced := suite.formatter.ReplaceLinks(context.Background(), text4)
	assert.Equal(suite.T(), `
two similar links:

<a href="https://example.org" rel="noopener">example.org</a>

<a href="https://example.org/test" rel="noopener">example.org/test</a>
`, replaced)
}

func (suite *LinkTestSuite) TestReplaceLinksFromText5() {
	// we know this one doesn't work properly, which is why html should always be sanitized before being passed into the ReplaceLinks function
	replaced := suite.formatter.ReplaceLinks(context.Background(), text5)
	assert.Equal(suite.T(), `
what happens when we already have a link within an href?

<a href="<a href="https://example.org" rel="noopener">example.org</a>"><a href="https://example.org" rel="noopener">example.org</a></a>
`, replaced)
}

func TestLinkTestSuite(t *testing.T) {
	suite.Run(t, new(LinkTestSuite))
}
