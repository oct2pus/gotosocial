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

package web

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

type Module struct {
	config    *config.Config
	processor processing.Processor
	log       *logrus.Logger
}

func New(config *config.Config, processor processing.Processor, log *logrus.Logger) api.ClientModule {
	return &Module{
		config:    config,
		log:       log,
		processor: processor,
	}
}

func (m *Module) baseHandler(c *gin.Context) {
	l := m.log.WithField("func", "BaseGETHandler")
	l.Trace("serving index html")

	instance, err := m.processor.InstanceGet(c.Request.Context(), m.config.Host)
	if err != nil {
		l.Debugf("error getting instance from processor: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// FIXME: fill in more variables?
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"instance":      instance,
		"countUsers":    3,
		"countStatuses": 42069,
		"version":       "1.0.0",
		"adminUsername": "@admin",
	})
}

func (m *Module) NotFoundHandler(c *gin.Context) {
	l := m.log.WithField("func", "404")
	l.Trace("serving 404 html")

	instance, err := m.processor.InstanceGet(c.Request.Context(), m.config.Host)
	if err != nil {
		l.Debugf("error getting instance from processor: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.HTML(404, "404.tmpl", gin.H{
		"instance": instance,
	})
}

// Route satisfies the RESTAPIModule interface
func (m *Module) Route(s router.Router) error {

	// serve static files from /assets
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %s", err)
	}
	assetPath := filepath.Join(cwd, m.config.TemplateConfig.AssetBaseDir)
	s.AttachStaticFS("/assets", FileSystem{http.Dir(assetPath)})

	// Admin panel route, if it exists
	adminPath := filepath.Join(cwd, m.config.TemplateConfig.AssetBaseDir, "/admin")
	s.AttachStaticFS("/admin", FileSystem{http.Dir(adminPath)})

	// serve front-page
	s.AttachHandler(http.MethodGet, "/", m.baseHandler)

	// 404 handler
	s.AttachNoRouteHandler(m.NotFoundHandler)

	return nil
}
