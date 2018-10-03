package main

import (
	"fmt"
	"log"

	redigo "github.com/garyburd/redigo/redis"
)

func setRedis(key string, value string) error {
	con := redisPool.Get()
	defer con.Close()

	_, err := con.Do("SET", key, value)
	// _, err = con.Do("EXPIRE", key, 1000)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return err
}

func incRedis(key string, value string) error {

	con := redisPool.Get()
	defer con.Close()

	_, err = con.Do("PING")
	if err != nil {
		log.Println("[redis] | Can't connect to the Redis database when INCR")
	}

	_, err := con.Do("INCR", key)
	// _, err = con.Do("EXPIRE", key, 1000)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return err
}

func GetRedis(key string) (string, error) {

	con := redisPool.Get()
	defer con.Close()

	_, err = con.Do("PING")
	if err != nil {
		log.Println("[redis] | Can't connect to the Redis database When GET VALUE")
	}

	return redigo.String(con.Do("GET", key))
}

// AddRedisCountLoadPage to update the num of count page accessed
func AddRedisCountLoadPage() error {
	key := fmt.Sprintf("training_db:%s", "hendrap")
	if value, err := GetRedis(key); err != nil {
		log.Println(err)
		log.Println("then, Add New Into Redis")
		value := "1"
		setRedis(key, value)
	} else {
		incRedis(key, value)
		return err
	}

	return err
}
