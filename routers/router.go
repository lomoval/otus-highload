package routers

import (
	"app/controllers"
	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/login", &controllers.LoginController{})
	beego.Router("/registration", &controllers.RegistrationController{})
	beego.Router("/friends", &controllers.FriendsController{})
	// beego.Router("/profile", &controllers.ProfileController{})
	beego.Router("/users", &controllers.UsersController{})
	beego.Router("/logout", &controllers.LogoutController{})

	beego.Include(&controllers.ProfileController{})
}
