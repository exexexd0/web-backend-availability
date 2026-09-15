package repository

import (
	"fmt"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type ConfigType string

const (
	ConfigTypeSingle      ConfigType = "Single Node"
	ConfigTypeClustering  ConfigType = "Clustering"
	ConfigTypeReplication ConfigType = "Replication"
)

type Status string

const (
	StatusPublished Status = "Published"
	StatusDraft     Status = "Draft"
	StatusDeleted   Status = "Deleted"
)

func (s Status) Label() string {
	switch s {
	case StatusPublished:
		return "Опубликован"
	case StatusDraft:
		return "Черновик"
	case StatusDeleted:
		return "Удалён"
	default:
		return string(s)
	}
}

type Component struct {
	ID               int
	Name             string
	Description      string
	ShortDescription string
	ConfigType       ConfigType
	UptimePercent    float32
	SystemImpact     float32
	Status           Status
	ImageURL         string
	VideoURL         string
	Likes            []int
}

func (r *Repository) GetComponent(id int) (*Component, error) {
	components, err := r.GetComponents()
	if err != nil {
		return nil, err
	}

	for i := range components {
		if components[i].ID == id {
			return &components[i], nil
		}
	}

	return nil, fmt.Errorf("компонент не найден")
}

func (r *Repository) GetDraftComponent() (*Component, error) {
	components, err := r.GetComponents()
	if err != nil {
		return nil, err
	}

	for i := range components {
		if components[i].Status == StatusDraft {
			return &components[i], nil
		}
	}

	return nil, fmt.Errorf("черновик не найден")
}

func (r *Repository) GetComponents() ([]Component, error) {
	// имитируем работу с БД. Типа мы выполнили sql запрос и получили эти строки из БД
	components := []Component{
		{
			ID:               1,
			Name:             "Кластер базы данных высокой доступности",
			Description:      "Активно-активный кластер из нескольких узлов обеспечивает работу без простоев во время обслуживания и автоматически переключается при отказе узла.",
			ShortDescription: "Многоузловая актив-актив архитектура кластера для работы без простоев...",
			ConfigType:       ConfigTypeClustering,
			UptimePercent:    99.99,
			SystemImpact:     60,
			Status:           StatusPublished,
			ImageURL:         "http://localhost:9000/media/%D0%BF%D0%BB%D0%B8%D1%82%D0%BA%D0%B01.jpg",
			VideoURL:         "http://localhost:9000/media/%D0%B2%D0%B8%D0%B4%D0%B5%D0%BE1.mp4",
			Likes:            []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			ID:               2,
			Name:             "Узел веб-сервера Альфа",
			Description:      "Одноузловой периферийный веб-уровень используется как базовая точка для сравнения доступности после репликации.",
			ShortDescription: "Одноузловой периферийный веб-уровень для сравнения доступности...",
			ConfigType:       ConfigTypeSingle,
			UptimePercent:    99.99,
			SystemImpact:     21.50,
			Status:           StatusPublished,
			ImageURL:         "http://localhost:9000/media/%D0%BF%D0%BB%D0%B8%D1%82%D0%BA%D0%B02.jpg",
			VideoURL:         "http://localhost:9000/media/%D0%B2%D0%B8%D0%B4%D0%B5%D0%BE%203.mp4",
			Likes:            []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			ID:               3,
			Name:             "Глобальный балансировщик нагрузки",
			Description:      "Балансировщик Anycast между региональными репликами сохраняет доступность трафика при частичных сбоях.",
			ShortDescription: "Балансировщик Anycast между региональными репликами...",
			ConfigType:       ConfigTypeReplication,
			UptimePercent:    99.95,
			SystemImpact:     10,
			Status:           StatusPublished,
			ImageURL:         "http://localhost:9000/media/%D0%BF%D0%BB%D0%B8%D1%82%D0%BA%D0%B03.jpg",
			VideoURL:         "http://localhost:9000/media/%D0%B2%D0%B8%D0%B4%D0%B5%D0%BE%203.mp4",
			Likes:            []int{1, 2, 4, 7},
		},
		{
			ID:               4,
			Name:             "Основная реплика базы данных",
			Description:      "Синхронная пара основной базы и реплики с почти нулевым RPO для контролируемых тренировок переключения.",
			ShortDescription: "Синхронная пара основной базы и реплики с почти нулевым RPO...",
			ConfigType:       ConfigTypeReplication,
			UptimePercent:    99.91,
			SystemImpact:     71.2,
			Status:           StatusPublished,
			ImageURL:         "http://localhost:9000/media/%D0%BF%D0%BB%D0%B8%D1%82%D0%BA%D0%B04.jpg",
			VideoURL:         "http://localhost:9000/media/%D0%B2%D0%B8%D0%B4%D0%B5%D0%BE%204.mp4",
			Likes:            []int{1},
		},
		{
			ID:               5,
			Name:             "Периферийный слой кэширования",
			Description:      "Уровень кэширования CDN снижает нагрузку на источник и скрывает кратковременные сбои источника.",
			ShortDescription: "Уровень кэширования CDN скрывает кратковременные сбои источника...",
			ConfigType:       ConfigTypeClustering,
			UptimePercent:    99.88,
			SystemImpact:     16,
			Status:           StatusPublished,
			ImageURL:         "http://localhost:9000/media/%D0%BF%D0%BB%D0%B8%D1%82%D0%BA%D0%B05.jpg",
			VideoURL:         "http://localhost:9000/media/%D0%B2%D0%B8%D0%B4%D0%B5%D0%BE%205.mp4",
			Likes:            []int{},
		},
		{
			ID:               6,
			Name:             "Агент мониторинга",
			Description:      "Черновой одноузловой агент проверок подготовлен для следующего сценария оценки доступности.",
			ShortDescription: "Черновой агент проверок для следующего сценария оценки...",
			ConfigType:       ConfigTypeSingle,
			UptimePercent:    99.50,
			SystemImpact:     89.11,
			Status:           StatusDraft,
			ImageURL:         "http://localhost:9000/media/%D0%BF%D0%BB%D0%B8%D1%82%D0%BA%D0%B06.jpg",
			VideoURL:         "http://localhost:9000/media/%D0%B2%D0%B8%D0%B4%D0%B5%D0%BE%206.mp4",
			Likes:            []int{1, 2, 3, 4, 5},
		},
		{
			ID:               7,
			Name:             "Устаревший брокер сообщений",
			Description:      "Устаревший брокер сохранён только для истории и не должен отображаться в интерфейсе.",
			ShortDescription: "Устаревший брокер скрыт из интерфейса...",
			ConfigType:       ConfigTypeSingle,
			UptimePercent:    97.10,
			SystemImpact:     23.00,
			Status:           StatusDeleted,
			ImageURL:         "http://localhost:9000/media/%D0%BF%D0%BB%D0%B8%D1%82%D0%BA%D0%B07.jpg",
			VideoURL:         "http://localhost:9000/media/%D0%B2%D0%B8%D0%B4%D0%B5%D0%BE%206.mp4",
			Likes:            []int{1, 2},
		},
	}

	if len(components) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return components, nil
}
