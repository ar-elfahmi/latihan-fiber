package database

import (
    "context"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
    "go.mongodb.org/mongo-driver/v2/mongo/readpref"
    
    "latihan-fiber/config"
)


// 1. buat opts dari URI di .env
func NewMongoClient(ctx context.Context) (*mongo.Client, error) { 
    opts := options.Client(). 
        ApplyURI(config.GetWithDefault("MONGO_URI", "mongodb://localhost:27017")). 
        SetServerSelectionTimeout(5 * time.Second). 
        SetMaxPoolSize(uint64(20)) 
  
// 2. connect
    client, err := mongo.Connect(opts) 
    if err != nil { 
        return nil, fmt.Errorf("menyiapkan client MongoDB: %w", err) 
    } 
  
// 3. ping dengan timeout 5 detik
    pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second) 
    defer cancel() 
    if err := client.Ping(pingCtx, readpref.Primary()); err != nil { 
        _ = client.Disconnect(context.Background()) 
        return nil, fmt.Errorf("ping MongoDB: %w", err) 
    } 
  
// 4. return client
    return client, nil 
}
