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

import (
	"net"
	"time"
)

// User represents an actual human user of gotosocial. Note, this is a LOCAL gotosocial user, not a remote account.
// To cross reference this local user with their account (which can be local or remote), use the AccountID field.
type User struct {
	/*
		BASIC INFO
	*/

	// id of this user in the local database; the end-user will never need to know this, it's strictly internal
	ID string `bun:"type:CHAR(26),pk,notnull,unique"`
	// confirmed email address for this user, this should be unique -- only one email address registered per instance, multiple users per email are not supported
	Email string `bun:"default:null,unique,nullzero"`
	// The id of the local gtsmodel.Account entry for this user, if it exists (unconfirmed users don't have an account yet)
	AccountID string   `bun:"type:CHAR(26),unique,nullzero"`
	Account   *Account `bun:"rel:belongs-to"`
	// The encrypted password of this user, generated using https://pkg.go.dev/golang.org/x/crypto/bcrypt#GenerateFromPassword. A salt is included so we're safe against 🌈 tables
	EncryptedPassword string `bun:",notnull"`

	/*
		USER METADATA
	*/

	// When was this user created?
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// From what IP was this user created?
	SignUpIP net.IP `bun:",nullzero"`
	// When was this user updated (eg., password changed, email address changed)?
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// When did this user sign in for their current session?
	CurrentSignInAt time.Time `bun:",nullzero"`
	// What's the most recent IP of this user
	CurrentSignInIP net.IP `bun:",nullzero"`
	// When did this user last sign in?
	LastSignInAt time.Time `bun:",nullzero"`
	// What's the previous IP of this user?
	LastSignInIP net.IP `bun:",nullzero"`
	// How many times has this user signed in?
	SignInCount int
	// id of the user who invited this user (who let this guy in?)
	InviteID string `bun:"type:CHAR(26),nullzero"`
	// What languages does this user want to see?
	ChosenLanguages []string
	// What languages does this user not want to see?
	FilteredLanguages []string
	// In what timezone/locale is this user located?
	Locale string `bun:",nullzero"`
	// Which application id created this user? See gtsmodel.Application
	CreatedByApplicationID string       `bun:"type:CHAR(26),nullzero"`
	CreatedByApplication   *Application `bun:"rel:belongs-to"`
	// When did we last contact this user
	LastEmailedAt time.Time `bun:",nullzero"`

	/*
		USER CONFIRMATION
	*/

	// What confirmation token did we send this user/what are we expecting back?
	ConfirmationToken string `bun:",nullzero"`
	// When did the user confirm their email address
	ConfirmedAt time.Time `bun:",nullzero"`
	// When did we send email confirmation to this user?
	ConfirmationSentAt time.Time `bun:",nullzero"`
	// Email address that hasn't yet been confirmed
	UnconfirmedEmail string `bun:",nullzero"`

	/*
		ACL FLAGS
	*/

	// Is this user a moderator?
	Moderator bool
	// Is this user an admin?
	Admin bool
	// Is this user disabled from posting?
	Disabled bool
	// Has this user been approved by a moderator?
	Approved bool

	/*
		USER SECURITY
	*/

	// The generated token that the user can use to reset their password
	ResetPasswordToken string `bun:",nullzero"`
	// When did we email the user their reset-password email?
	ResetPasswordSentAt time.Time `bun:",nullzero"`

	EncryptedOTPSecret     string `bun:",nullzero"`
	EncryptedOTPSecretIv   string `bun:",nullzero"`
	EncryptedOTPSecretSalt string `bun:",nullzero"`
	OTPRequiredForLogin    bool
	OTPBackupCodes         []string
	ConsumedTimestamp      int
	RememberToken          string    `bun:",nullzero"`
	SignInToken            string    `bun:",nullzero"`
	SignInTokenSentAt      time.Time `bun:",nullzero"`
	WebauthnID             string    `bun:",nullzero"`
}
