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

package federatingdb

import (
	"context"
	"net/url"

	"github.com/sirupsen/logrus"
)

// Exists returns true if the database has an entry for the specified
// id. It may not be owned by this application instance.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Exists(c context.Context, id *url.URL) (exists bool, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "Exists",
			"id":   id.String(),
		},
	)
	l.Debugf("entering EXISTS function with id %s", id.String())

	return false, nil
}
