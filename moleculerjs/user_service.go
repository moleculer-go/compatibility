package moleculerjs

import (
	"errors"

	"github.com/moleculer-go/moleculer"
)

type UserService struct {
	profileCreated chan bool
}

func (s *UserService) Name() string {
	return "user"
}

func (s *UserService) Dependencies() []string {
	return []string{"profile"}
}

func (s *UserService) Create(ctx moleculer.Context, user moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.create called! - user: ", user)
	ctx.Emit("user.created", user)
	return user
}

func (s *UserService) Get(ctx moleculer.Context, user moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.get called! - user: ", user)
	return user
}

func (s *UserService) Update(ctx moleculer.Context, user moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.update called! - user: ", user)
	ctx.Emit("user.updated", user)
	return user
}

func (s *UserService) Panix(ctx moleculer.Context, params moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.panix called! ")
	panic("this action will panic!")
}

func (s *UserService) Fail(ctx moleculer.Context) interface{} {
	ctx.Logger().Info("user.fail called! ")
	return errors.New("this actions returns an error!")
}

func (s *UserService) Events() []moleculer.Event {
	return []moleculer.Event{
		{
			Name: "profile.loopevent",
			Handler: func(ctx moleculer.Context, params moleculer.Payload) {
				ctx.Logger().Info("profile.loopevent arrived: ", params)
			},
		},
		{
			Name: "profile.created",
			Handler: func(ctx moleculer.Context, profile moleculer.Payload) {
				ctx.Logger().Info("profile.created event! profile: ", profile)
				user := map[string]interface{}{
					"id":        profile.Get("user").Get("id").String(),
					"profileId": profile.Get("id").String(),
				}
				<-ctx.Call("user.update", user)
				ctx.Logger().Info("user updated with profile Id :) ")

				go func() {
					s.profileCreated <- true
				}()
			},
		},
	}
}
