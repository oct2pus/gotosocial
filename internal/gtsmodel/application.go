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

// Application represents an application that can perform actions on behalf of a user.
// It is used to authorize tokens etc, and is associated with an oauth client id in the database.
type Application struct {
	// id of this application in the db
	ID string `bun:"type:CHAR(26),pk,notnull"`
	// name of the application given when it was created (eg., 'tusky')
	Name string `bun:",nullzero"`
	// website for the application given when it was created (eg., 'https://tusky.app')
	Website string `bun:",nullzero"`
	// redirect uri requested by the application for oauth2 flow
	RedirectURI string `bun:",nullzero"`
	// id of the associated oauth client entity in the db
	ClientID string `bun:"type:CHAR(26),nullzero"`
	// secret of the associated oauth client entity in the db
	ClientSecret string `bun:",nullzero"`
	// scopes requested when this app was created
	Scopes string `bun:",nullzero"`
	// a vapid key generated for this app when it was created
	VapidKey string `bun:",nullzero"`
}
