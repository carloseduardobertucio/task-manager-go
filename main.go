package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	gin "github.com/gin-gonic/gin"
)

type Task struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    string    `json:"priority"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	Completed   bool      `json:"completed"`
}

type TaskList struct {
	Tasks []Task `json:"tasks"`
}

func loadTasks() TaskList {
	file, err := os.ReadFile("tasks.json")
	if err != nil {
		log.Printf("Error reading tasks file: %v", err)
		return TaskList{}
	}

	var taskList TaskList
	err = json.Unmarshal(file, &taskList)
	if err != nil {
		log.Printf("Error unmarshaling tasks: %v", err)
		return TaskList{}
	}

	return taskList
}

func saveTasks(tasks TaskList) error {
	data, err := json.MarshalIndent(tasks, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile("tasks.json", data, 0644)
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		tasks := loadTasks()
		c.HTML(http.StatusOK, "index.html", gin.H{
			"tasks": tasks.Tasks,
		})
	})

	r.POST("/add", func(c *gin.Context) {
		task := Task{
			Title:       c.PostForm("title"),
			Description: c.PostForm("description"),
			Priority:    c.PostForm("priority"),
			Category:    c.PostForm("category"),
			CreatedAt:   time.Now(),
			Completed:   false,
		}

		tasks := loadTasks()
		tasks.Tasks = append(tasks.Tasks, task)

		if err := saveTasks(tasks); err != nil {
			c.String(http.StatusInternalServerError, "Error saving task")
			return
		}

		c.Redirect(http.StatusFound, "/")
	})

	r.GET("/complete/:index", func(c *gin.Context) {
		index := c.Param("index")
		tasks := loadTasks()

		var idx int
		_, err := fmt.Sscanf(index, "%d", &idx)
		if err != nil || idx < 0 || idx >= len(tasks.Tasks) {
			c.String(http.StatusBadRequest, "Invalid index")
			return
		}

		tasks.Tasks[idx].Completed = true
		if err := saveTasks(tasks); err != nil {
			c.String(http.StatusInternalServerError, "Error updating task")
			return
		}

		c.Redirect(http.StatusFound, "/")
	})

	r.GET("/delete/:index", func(c *gin.Context) {
		index := c.Param("index")
		tasks := loadTasks()

		var idx int
		_, err := fmt.Sscanf(index, "%d", &idx)
		if err != nil || idx < 0 || idx >= len(tasks.Tasks) {
			c.String(http.StatusBadRequest, "Invalid index")
			return
		}

		tasks.Tasks = append(tasks.Tasks[:idx], tasks.Tasks[idx+1:]...)
		if err := saveTasks(tasks); err != nil {
			c.String(http.StatusInternalServerError, "Error deleting task")
			return
		}

		c.Redirect(http.StatusFound, "/")
	})

	r.Run(":8080")
}
