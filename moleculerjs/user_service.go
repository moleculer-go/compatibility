package moleculerjs

import (
	"github.com/moleculer-go/moleculer"
)

type UserService struct {
	profileCreated chan bool
}

func (s *UserService) Name() string {
	return "user"
}

func (s *UserService) Create(ctx moleculer.Context, user moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.create called! - user: ", user)
	ctx.Emit("user.created", user)
	return user
}

func (s *UserService) Update(ctx moleculer.Context, user moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.update called! - user: ", user)
	ctx.Emit("user.updated", user)
	return user
}

func (s *UserService) Events() []moleculer.Event {
	return []moleculer.Event{
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

				s.profileCreated <- true
			},
		},
	}
}
