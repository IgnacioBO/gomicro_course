package course

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/IgnacioBO/gomicro_domain/domain"
)

type Repository interface {
	Create(ctx context.Context, c *domain.Course) error                                      //Metodo create y recibe un Puntero de un Course (Struct creado en el de domain.go, que tiene los campso de BBDD en gorn)
	GetAll(ctx context.Context, filtros Filtros, offset, limit int) ([]domain.Course, error) //Le agregamos que getAll reciba filtros
	Get(ctx context.Context, id string) (*domain.Course, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, id string, name *string, startDate, endDate *time.Time) error //Campos por separado y como punteros (porque si no lo pongo puntero, si llega un string vacio TENDRA valor y actualizará VACIO)
	Count(ctx context.Context, Filtros Filtros) (int, error)                                  //Servirá para contar cantidad de registrosy recibe los mismo filtros del getall y devolera int(cantidad de registros) y error
}

type repo struct {
	log *log.Logger
	db  *gorm.DB
}

func NewRepo(log *log.Logger, db *gorm.DB) Repository {
	return &repo{
		log: log,
		db:  db, //Devuevle un struct repo con la bbdd
	}

}

func (r *repo) Create(ctx context.Context, course *domain.Course) error {
	r.log.Println("repository Create:", course)

	result := r.db.WithContext(ctx).Create(course)

	if result.Error != nil {
		r.log.Println(result.Error)
		return result.Error
	}
	r.log.Printf("course created with id: %s, rows affected: %d\n", course.ID, result.RowsAffected)
	return nil
}

func (r *repo) GetAll(ctx context.Context, filtros Filtros, offset, limit int) ([]domain.Course, error) {
	r.log.Println("repository GetAll:")

	var allCourses []domain.Course

	tx := r.db.WithContext(ctx).Model(&allCourses)
	tx = aplicarFiltros(tx, filtros)
	tx = tx.Limit(limit).Offset(offset)
	result := tx.Order("created_at desc").Find(&allCourses)
	if result.Error != nil {
		r.log.Println(result.Error)
		return nil, result.Error
	}
	r.log.Printf("all courses retrieved, rows affected: %d\n", result.RowsAffected)
	return allCourses, nil
}

func (r *repo) Get(ctx context.Context, id string) (*domain.Course, error) {
	r.log.Println("repository Get by id:")

	usuario := domain.Course{ID: id}

	result := r.db.WithContext(ctx).First(&usuario)
	if result.Error != nil {
		r.log.Println(result.Error)
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrCourseNotFound{id}
		}
		return nil, result.Error
	}
	r.log.Printf("course retrieved with id: %s, rows affected: %d\n", id, result.RowsAffected)
	return &usuario, nil
}

func (r *repo) Delete(ctx context.Context, id string) error {
	r.log.Println("repository Delete by id:")

	usuario := domain.Course{ID: id}

	result := r.db.WithContext(ctx).Delete(&usuario)
	if result.Error != nil {
		r.log.Println(result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		r.log.Println("%s", ErrCourseNotFound{id}.Error(), result.RowsAffected)
		return ErrCourseNotFound{id}
	}
	r.log.Printf("course deleted with id: %s, rows affected: %d\n", id, result.RowsAffected)
	return nil
}

func (r *repo) Update(ctx context.Context, id string, name *string, startDate, endDate *time.Time) error {
	r.log.Println("repository Update")

	valores := make(map[string]interface{})

	if name != nil {
		valores["name"] = *name
	}

	if startDate != nil {
		valores["start_date"] = *startDate
	}

	if endDate != nil {
		valores["end_date"] = *endDate
	}

	result := r.db.WithContext(ctx).Model(domain.Course{}).Where("id = ?", id).Updates(valores)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		r.log.Println("%s", ErrCourseNotFound{id}.Error(), result.RowsAffected)
		return ErrCourseNotFound{id}
	}
	r.log.Printf("course updated with id: %s, rows affected: %d\n", id, result.RowsAffected)

	return nil
}

// Funcion que servira para filtrar, recibe la base da datos (tx) y el struct de filtros
func aplicarFiltros(tx *gorm.DB, filtros Filtros) *gorm.DB {
	//Si el filtro es distinto de blanco (osea VIENE con filtro), le agregaremos un fultros
	if filtros.Name != "" {

		filtros.Name = fmt.Sprintf("%%%s%%", strings.ToLower(filtros.Name))
		//El Where filtra el valor que le paso, osea el Where permite AGREGAR un Where a la consulta
		tx = tx.Where("lower(name) like ?", filtros.Name)
	}
	return tx
}

func (r *repo) Count(ctx context.Context, filtros Filtros) (int, error) {
	var cantidad int64
	tx := r.db.WithContext(ctx).Model(domain.Course{})
	tx = aplicarFiltros(tx, filtros)
	tx = tx.Count(&cantidad)
	if tx.Error != nil {
		return 0, tx.Error
	}

	return int(cantidad), nil
}
