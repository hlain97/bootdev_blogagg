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
"encoding/xml"
"io"
"net/http"
"html"

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

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed (ctx context.Context, feedURL string)(*RSSFeed, error){
	newFeed := &RSSFeed{}
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)

	if err != nil{
		return nil, err
	}

	req.Header.Add("User-Agent", "gator")

	res, err := http.DefaultClient.Do(req)

	if err != nil{
		return nil, err
	}

	defer res.Body.Close()

	read, err := io.ReadAll(res.Body)

	if err != nil{
		return nil, err
	}

	err = xml.Unmarshal(read, newFeed)

	if err != nil{
		return nil, err
	}

	for i := range newFeed.Channel.Item {
		newFeed.Channel.Item[i].Title = html.UnescapeString(newFeed.Channel.Item[i].Title)
		newFeed.Channel.Item[i].Description = html.UnescapeString(newFeed.Channel.Item[i].Description)
	}

	newFeed.Channel.Title = html.UnescapeString(newFeed.Channel.Title)
	newFeed.Channel.Description = html.UnescapeString(newFeed.Channel.Description)


	return newFeed, nil
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

func handlerFeeds(s *state, cmd command) error {

	dbs, err := s.db.GetFeeds(context.Background())

	if err != nil {
		return err
	}

	for _, feed := range dbs {
		fmt.Println(feed.Name)
		fmt.Println(feed.Url)
		fmt.Println(feed.UserName)
	}

	return nil

}

func handlerFeed(s *state, cmd command) error{
	if len(cmd.Arguments) < 2 {
		return fmt.Errorf("No name provided")
	}

	user, err := s.db.GetUser(context.Background(), s.Config.CurrentUserName)
	if err != nil {
    return err
}

	name := cmd.Arguments[0]

	params := database.CreateFeedParams{
		ID:	uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt:	time.Now(),
		Name: name,
		Url: cmd.Arguments[1],
		UserID: user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), params)

	if err != nil {
		return err
	}

	fmt.Println(feed)
	return nil

}


func handlerAgg(s *state, cmd command) error {
	feed, err := fetchFeed(context.Background(),"https://www.wagslane.dev/index.xml")

	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", feed)

	return nil
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
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", handlerFeed)
	cmds.register("feeds", handlerFeeds)

	err = cmds.run(&s, cmd)

	if err != nil {
		log.Fatal(err)
	}


}