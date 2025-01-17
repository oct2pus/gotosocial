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

package gtsmodel

import "time"

// Tag represents a hashtag for gathering public statuses together
type Tag struct {
	// id of this tag in the database
	ID string `bun:",unique,type:CHAR(26),pk,notnull"`
	// Href of this tag, eg https://example.org/tags/somehashtag
	URL string `bun:",nullzero"`
	// name of this tag -- the tag without the hash part
	Name string `bun:",unique,notnull"`
	// Which account ID is the first one we saw using this tag?
	FirstSeenFromAccountID string `bun:"type:CHAR(26),nullzero"`
	// when was this tag created
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// when was this tag last updated
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// can our instance users use this tag?
	Useable bool `bun:",notnull,default:true"`
	// can our instance users look up this tag?
	Listable bool `bun:",notnull,default:true"`
	// when was this tag last used?
	LastStatusAt time.Time `bun:",nullzero"`
}
