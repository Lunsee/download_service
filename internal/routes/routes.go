package routes

import (
	"download_service/internal/models"
	"download_service/internal/service"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/mux"

	"github.com/google/uuid"
)

var taskMap = make(map[uuid.UUID][]models.Task)
var (
	taskMapMu sync.Mutex

	MAX_LINKS_PER_TASK = 3
	MAX_USER_TASKS     = 3
)

func GenerateID() uuid.UUID {
	return uuid.New()
}

// CreateTask godoc
// @Summary Создать новую задачу
// @Description Создает новую задачу для загрузки файлов
// @Tags tasks
// @Accept json
// @Produce json
// @Param data body object false "Данные запроса" example={"user_id":"97557152-c723-44c8-bb92-241d37a81344"}
// @Success 201 {object} object "Успешный ответ" example={"user_id":"97557152-c723-44c8-bb92-241d37a81344","task_id":"bcdfcfc5-f1ab-4118-a5a6-681f69df1698"}
// @Failure 400 {object} object "Ошибка" example={"message":"error decoding JSON"}
// @Failure 404 {object} object "Ошибка" example={"message":"user not found"}
// @Router /createTask [post]
func CreateTask(w http.ResponseWriter, r *http.Request) {
	log.Printf("create task endpoint..")

	bodyBytes, _ := io.ReadAll(r.Body)
	if len(bodyBytes) == 0 || string(bodyBytes) == "null" {
		log.Printf("request body empty...")
		newUserID := GenerateID()

		initialTask := models.Task{
			ID:     GenerateID(),
			Status: models.StatusPending,
			Links:  []string{},
			Errors: []string{},
		}
		taskMap[newUserID] = []models.Task{initialTask}

		log.Printf("new user created with ID: %s and initial empty task", newUserID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"user_id": newUserID.String(),
			"task_id": initialTask.ID.String(),
		})
		return
	}

	var body struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		http.Error(w, "error decoding JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	parsedUUID, err := uuid.Parse(body.UserID)
	if err != nil {
		log.Fatalf("Invalid UUID: %v", err)
	}

	if parsedUUID == uuid.Nil {
		http.Error(w, "user_id cannot be empty", http.StatusBadRequest)
		return
	}
	log.Printf("user id want to create task: %s", parsedUUID)

	userTasks, exists := taskMap[parsedUUID]
	if !exists {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if len(userTasks) >= MAX_USER_TASKS {
		http.Error(w, "maximum 3 tasks per user allowed", http.StatusBadRequest)
		return
	}

	newTask := models.Task{
		ID:     GenerateID(),
		Status: models.StatusPending,
		Links:  []string{},
		Errors: []string{},
	}
	taskMap[parsedUUID] = append(taskMap[parsedUUID], newTask)

	log.Printf("new empty task created for user %s", body.UserID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"user_id": parsedUUID.String(),
		"task_id": newTask.ID.String(),
	})
}

// AddTaskItems godoc
// @Summary Добавить ссылки в задачу
// @Description Добавляет список ссылок для загрузки в указанную задачу
// @Tags tasks
// @Accept json
// @Produce json
// @Param data body object true "Данные для добавления" example={"user_id":"97557152-c723-44c8-bb92-241d37a81344","task_id":"bcdfcfc5-f1ab-4118-a5a6-681f69df1698","links":["https://example.com/file1.pdf","https://example.com/image.jpg"]}
// @Success 200 {object} object "Обновленная задача" example={"id":"bcdfcfc5-f1ab-4118-a5a6-681f69df1698","status":"working","links":["https://example.com/file1.pdf","https://example.com/image.jpg"],"errors":[]}
// @Failure 400 {object} object "Ошибка" example={"message":"error decoding JSON"}
// @Failure 404 {object} object "Ошибка" example={"message":"task not found"}
// @Failure 500 {object} object "Ошибка" example={"message":"Server is busy"}
// @Router /addTaskItems [post]
func AddTaskItems(w http.ResponseWriter, r *http.Request) {
	log.Printf("edit task endpoint..")

	type request struct {
		UserID string   `json:"user_id"`
		TaskID string   `json:"task_id"`
		Links  []string `json:"links"`
	}

	if r.Body == nil {
		http.Error(w, "error: Empty body request", http.StatusBadRequest)
		return
	}

	var req request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "error decoding JSON", http.StatusBadRequest)
		return
	}

	//conv to uuid
	parsedUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		log.Fatalf("Invalid UUID: %v", err)
		return
	}

	if parsedUserID == uuid.Nil {
		http.Error(w, "user_id cannot be empty", http.StatusBadRequest)
		return
	}

	parsedTaskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		log.Fatalf("Invalid UUID: %v", err)
		return
	}

	if parsedTaskID == uuid.Nil {
		http.Error(w, "user_id cannot be empty", http.StatusBadRequest)
		return
	}

	log.Printf(" user id: %s want to add items task: %s", parsedUserID, parsedTaskID)

	taskMapMu.Lock()
	defer taskMapMu.Unlock()

	// find user and task id
	userTasks, exists := taskMap[parsedUserID]
	if !exists {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	var task *models.Task

	for i := range userTasks {
		if userTasks[i].ID == parsedTaskID {
			task = &userTasks[i]
			break
		}
	}

	if task == nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	//check count links
	if len(task.Links)+len(req.Links) > MAX_LINKS_PER_TASK {
		http.Error(w, "Server is busy", http.StatusBadRequest)
		return
	}

	task.Mu.Lock()
	task.Links = append(task.Links, req.Links...)

	//set task status
	if task.Status != models.StatusWorking && len(task.Links) > 0 {
		task.Status = models.StatusWorking
	}
	task.Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// GetTaskStatus godoc
// @Summary Получить статус задачи
// @Description Возвращает текущий статус задачи. Если задача завершена, возвращает URL для скачивания.
// @Tags tasks
// @Accept json
// @Produce json
// @Param data body object true "Данные запроса" example={"user_id":"97557152-c723-44c8-bb92-241d37a81344","task_id":"bcdfcfc5-f1ab-4118-a5a6-681f69df1698"}
// @Success 200 {object} object "Статус задачи" example={"status":"working"}
// @Success 200 {object} object "Завершенная задача" example={"status":"completed","download_url":"http://localhost:8080/download/archive.zip"}
// @Failure 400 {object} object "Ошибка" example={"message":"error decoding JSON"}
// @Failure 404 {object} object "Ошибка" example={"message":"task not found"}
// @Router /taskStatus [get]
func GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetTaskStatus endpoint..")
	type request struct {
		UserID string `json:"user_id"`
		TaskID string `json:"task_id"`
	}

	if r.Body == nil {
		http.Error(w, "error: Empty body request", http.StatusBadRequest)
		return
	}

	var req request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "error decoding JSON", http.StatusBadRequest)
		return
	}

	//conv to uuid
	parsedUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		log.Fatalf("Invalid UUID: %v", err)
		return
	}

	if parsedUserID == uuid.Nil {
		http.Error(w, "user_id cannot be empty", http.StatusBadRequest)
		return
	}

	parsedTaskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		log.Fatalf("Invalid UUID: %v", err)
		return
	}

	if parsedTaskID == uuid.Nil {
		http.Error(w, "user_id cannot be empty", http.StatusBadRequest)
		return
	}

	log.Printf(" user id: %s want to check status task: %s", parsedUserID, parsedTaskID)

	taskMapMu.Lock()
	defer taskMapMu.Unlock()

	// find user and task id
	userTasks, exists := taskMap[parsedUserID]
	if !exists {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	var task *models.Task

	for i := range userTasks {
		if userTasks[i].ID == parsedTaskID {
			task = &userTasks[i]
			break
		}
	}

	if task == nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	//check count links
	if len(task.Links) == MAX_LINKS_PER_TASK {
		download_url, err := service.Download(task, models.StorageDir, models.BaseUrl)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		task.Status = models.StatusComplete

		response := map[string]interface{}{
			"status":       task.Status,
			"download_url": download_url,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
		return
	}

	response := map[string]interface{}{
		"status": task.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// Download godoc
// @Summary Скачать архив с файлами
// @Description Позволяет скачать zip-архив с загруженными файлами по ID задачи
// @Tags files
// @Produce octet-stream
// @Param file_id path string true "ID архива (формат UUID с расширением .zip)" example:"d61fe9b2-0b1b-4b66-97b3-a5d46f02ee67.zip"
// @Success 200 {file} binary "Zip-архив с файлами"
// @Header 200 {string} Content-Disposition "attachment; filename=archive.zip"
// @Header 200 {string} Content-Type "application/zip"
// @Failure 400 {object} object "Неверный запрос" example={"message":"Invalid file type"}
// @Failure 404 {object} object "Файл не найден" example={"message":"File not found"}
// @Router /download/{file_id} [get]
func Download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["file_id"]

	if fileID == "" {
		http.Error(w, "File ID not specified", http.StatusBadRequest)
		return
	}

	// check name: .zip
	if !strings.HasSuffix(fileID, ".zip") {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}
	filePath := filepath.Join(models.StorageDir, fileID)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileID))

	http.ServeFile(w, r, filePath)
}
