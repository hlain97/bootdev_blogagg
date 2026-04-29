package main

import (
"fmt"
"log"
"errors"
"os"
"database/sql"
_ "github.com/lib/pq"
"time"
"context"

"github.com/google/uuid"
"github.com/hlain97/blog/internal/config"
"github.com/hlain97/blog/internal/database"
)

type commands struct {
	handlers map[string]func(*state, command) error

}

type state struct {
	db	*database.Queries
	Config *config.Config
}

type command struct {
	Name string
	Arguments []string
}

func (c *commands) run(s *state, cmd command) error{

	handler, ok := c.handlers[cmd.Name]

	if !ok {
		return fmt.Errorf("Command does not exist: %s", cmd.Name)
	}

	return handler(s, cmd)
	
}
func (c *commands) register(name string, f func(*state, command) error){
	c.handlers[name] = f
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Arguments) < 1 {
		return fmt.Errorf("No name provided")
	}

	name := cmd.Arguments[0]

	params := database.CreateUserParams{
		ID:	uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt:	time.Now(),
		Name: name,
	}

	user, err := s.db.CreateUser(context.Background(), params)

	if err != nil {
		return err
	}

	s.Config.SetUser(user.Name)
	fmt.Println("User created")
	fmt.Printf("User created: %+v\n", user)
	return nil
}


func handlerLogin(s *state, cmd command) error {
	if len(cmd.Arguments) == 0 {
		return errors.New("login requires a username argument")
	}

	username := cmd.Arguments[0]

	// Check if user exists in DB first
	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return errors.New("user does not exist")
	}

	// Only set config if valid user
	err = s.Config.SetUser(username)
	if err != nil {
		return err
	}

	fmt.Printf("Username set to %s\n", username)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("Database reset successfully")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Name == s.Config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
			continue
		}

		fmt.Printf("* %s\n", user.Name)
	}

	return nil
}

func main(){
	userInput := os.Args

	if len(userInput) < 2 {
		log.Fatal("Need more than one input")
	}

	cmd := command{
		Name: userInput[1],
		Arguments : userInput[2:],
	}

	cfg, err := config.Read()

	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
	log.Fatal(err)
	}
	defer db.Close()

	dbQueries := database.New(db)

	s := state{
		db: dbQueries,
		Config: &cfg,
	}

	cmds := commands{
		handlers: map[string]func(*state, command) error{},
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)

	err = cmds.run(&s, cmd)

	if err != nil {
		log.Fatal(err)
	}


}