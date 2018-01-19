package web

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	"github.com/ktt-ol/sesam/conf"
	"fmt"
	"github.com/ktt-ol/sesam/wikidata"
	"github.com/sirupsen/logrus"
	"strings"
	"github.com/ktt-ol/sesam/mqtt"
	"github.com/utrack/gin-csrf"
)

const KEY_USER_NAME = "userName"
const REMEMBER_PASSWORD_DAYS = 180;

var logger = logrus.WithField("where", "web")

type web struct {
	wikiData    *wikidata.WikiData
	mqttHandler *mqtt.MqttHandler
}

func StartWeb(config conf.ServerConf, wikiData *wikidata.WikiData, mqttHandler *mqtt.MqttHandler) {
	webHandler := web{wikiData, mqttHandler}

	keys := conf.GetKeys(config.KeysFile)

	gin.DisableConsoleColor()
	gin.DefaultWriter = logrus.WithField("where", "gin").WriterLevel(logrus.DebugLevel)
	gin.DefaultErrorWriter = logrus.WithField("where", "gin").WriterLevel(logrus.ErrorLevel)

	router := gin.Default()

	store := sessions.NewCookieStore(keys.SessionAuthKey, keys.SessionEncryptionKey)
	store.Options(sessions.Options{HttpOnly: true, Secure: true})
	router.Use(sessions.Sessions("sesam", store))

	router.Use(csrf.Middleware(csrf.Options{
		Secret: keys.CsrfKey,
		ErrorFunc: func(c *gin.Context){
			logger.Warn("CSRF token mismatch")
			c.String(400, "CSRF token mismatch")
			c.Abort()
		},
	}))

	router.Static("/assets", "webUI/assets")
	router.StaticFile("/swDummy.js", "webUI/swDummy.js")
	router.LoadHTMLGlob("webUI/templates/*.html")

	router.GET("/", webHandler.getMain)
	router.PUT("/buzzer", webHandler.putBuzzer)

	router.GET("/login", webHandler.getLogin)
	router.POST("/login", webHandler.postLogin)
	router.GET("/logout", webHandler.getLogout)

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	var err error
	if config.Https {
		err = router.RunTLS(addr, config.CertFile, config.CertKeyFile)
	} else {
		err = router.Run(addr)
	}
	if err != nil {
		logger.Error("gin exit", err)
	}
}

func (w *web) getMain(c *gin.Context) {
	session := sessions.Default(c)
	loginV := session.Get(KEY_USER_NAME)
	if loginV == nil {
		logger.Info("Not logged in.")
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}
	login := loginV.(string)

	isOpen := isOpenForMember(w.mqttHandler.CurrentStatus())
	var status string
	if isOpen {
		status = "opened"
	} else {
		status = "closed"
	}
	c.HTML(http.StatusOK, "index.html", gin.H{
		"login":       login,
		"statusClass": status,
		"isOpen":      isOpen,
		"csrf": csrf.GetToken(c),
	})
}

func (w *web) putBuzzer(c *gin.Context) {
	session := sessions.Default(c)
	loginV := session.Get(KEY_USER_NAME)
	if loginV == nil {
		logger.Info("Not logged in.")
		c.String(200, "LOGIN")
		return
	}
	userName := loginV.(string)

	doorStr := c.Query("door")
	var door mqtt.Door
	if doorStr == "inner" {
		door = mqtt.DoorInner
	} else if doorStr == "outer" {
		door = mqtt.DoorOuter
	} else {
		logger.WithField("doorStr", doorStr).Error("Invalid 'door' param")
		sendError(c, "Invalid 'door' param.")
		return
	}

	ok := w.mqttHandler.SendDoorBuzzer(door, userName)
	if ok {
		c.String(200, "OK")
	} else {
		c.String(200, "ERROR")
	}

}

func (w *web) getLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"days": REMEMBER_PASSWORD_DAYS,
		"csrf": csrf.GetToken(c),
	})
}

func (w *web) postLogin(c *gin.Context) {
	var form loginData
	if err := c.Bind(&form); err != nil {
		logger.WithError(err).Error("Invalid binding.")
		sendError(c, "Invalid binding.")
		return
	}

	form.Email = strings.ToLower(form.Email)
	success, userName, _ := w.wikiData.CheckPassword(form.Email, form.Password)
	if !success {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"days":  REMEMBER_PASSWORD_DAYS,
			"error": true,
			"csrf": csrf.GetToken(c),
		})
		return
	}

	session := sessions.Default(c)
	if len(form.Remember) > 0 {
		maxAgeSeconds := REMEMBER_PASSWORD_DAYS * 24 * 60 * 60
		session.Options(sessions.Options{MaxAge: maxAgeSeconds, HttpOnly: true, Secure: true})
	}
	session.Set(KEY_USER_NAME, userName)
	session.Save()

	c.Redirect(http.StatusSeeOther, "/")
}

func (w *web) getLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()

	c.Redirect(http.StatusSeeOther, "/login")
}

func sendError(c *gin.Context, msg string) {
	c.String(http.StatusBadRequest, "Error: "+msg)
	c.Abort()
}

// isOpenForMember returns true if the given textual status represents an open statue for normal member.
func isOpenForMember(mqttStatus string) bool {
	return mqttStatus == "open+" || mqttStatus == "open" || mqttStatus == "member"
}

type loginData struct {
	Email    string `form:"email" binding:"required"`
	Password string `form:"password" binding:"required"`
	Remember string `form:"remember"`
}
