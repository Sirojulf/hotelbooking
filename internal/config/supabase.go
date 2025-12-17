package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"github.com/supabase-community/supabase-go"
)

var SupabaseClient *supabase.Client

func ConnectSupabase() error {

	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No .env file found")
	}

	supaURL := viper.GetString("SUPABASE_URL")
	if supaURL == "" {
		return fmt.Errorf("SUPABASE_URL is not set")
	}

	supaKEY := viper.GetString("SUPABASE_KEY")
	if supaKEY == "" {
		return fmt.Errorf("SUPABASE_KEY is not set")
	}

	client, err := supabase.NewClient(supaURL, supaKEY, &supabase.ClientOptions{})
	if err != nil {
		return fmt.Errorf("failed to initialize the client: %w", err)
	}

	SupabaseClient = client
	fmt.Println("Success connected to Supabase.... ")

	return nil
}
