package course

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/IgnacioBO/gomicro_domain/domain"
)

type Service interface {
	Create(ctx context.Context, name, startDate, endDate string) (*domain.Course, error)     //Metodo que recibira datos de creacion y devolvera un error (y la entidad domain.Course)
	GetAll(ctx context.Context, filtros Filtros, offset, limit int) ([]domain.Course, error) //Le agregamos filtros (con el struct filtro sque creamos)
	Get(ctx context.Context, id string) (*domain.Course, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, id string, name *string, startDate, endDate *string) error
	Count(ctx context.Context, Filtros Filtros) (int, error) //Servirá para contar cantidad de registrosy recibe los mismo filtros del getall y devolera int(cantidad de registros) y error
}

// Ahora crearemos un struct PRIVADA (pq desde afuera accederemoa a traves de Servivce)
// Recibira un repository (de la capa repositry)
// Tambien recibira un logger
type service struct {
	log  *log.Logger
	repo Repository
}

// Crea (instanciar) un servicio que sera la interfaz (devovlerá una interface de tupo Service [creado arriba], PERO hara un RETURN especificamente del STRUCT service (con minusculas))
// Recibirá un objeo Repositor y devovlera un service con el repo
// Tambien recibira un logger
func NewService(log *log.Logger, repo Repository) Service {
	return &service{
		log:  log,
		repo: repo,
	}
}

type Filtros struct {
	Name string
}

func (s service) Create(ctx context.Context, name, startDate, endDate string) (*domain.Course, error) {
	s.log.Println("Create course service")

	//Si tienen texto con un T posterior, por ejemplo 2025-01-27T22:59:09.409Z, se saca
	if strings.Contains(startDate, "T") {
		startDate = strings.Split(startDate, "T")[0]
	}
	if strings.Contains(endDate, "T") {
		endDate = strings.Split(endDate, "T")[0]
	}

	//Parse para que transforme el startDate y enddate de string a fecha (osea llegara  xxxx-xx-xx y se trasnforara en time.Time)
	startDateParsed, err := time.Parse("2006-01-02", startDate) //Parse
	if err != nil {
		s.log.Println(err)
		return nil, err
	}
	endDateParsed, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		s.log.Println(err)
		return nil, err
	}

	cursoNuevo := domain.Course{
		Name:      name,
		StartDate: startDateParsed,
		EndDate:   endDateParsed,
	}
	//Le pasamo al repo el domain.Course (del domain.go) a la capa repo a la funcion Create (que recibe puntero)
	err = s.repo.Create(ctx, &cursoNuevo)
	//Si hay un error (por ejemplo al insertar, se devuelve el error y la capa endpoitn lo maneja con un status code y todo)
	if err != nil {
		return nil, err
	}
	return &cursoNuevo, nil
}

func (s service) GetAll(ctx context.Context, filtros Filtros, offset, limit int) ([]domain.Course, error) {
	s.log.Println("GetAll courses service")

	allCourses, err := s.repo.GetAll(ctx, filtros, offset, limit)
	if err != nil {
		return nil, err
	}
	//OJo aqui devuelve el start_data y end_date con horario loca (osea -3, aqui podriamos hacer algo para pasarlo a UTC)
	return allCourses, nil
}

func (s service) Get(ctx context.Context, id string) (*domain.Course, error) {
	s.log.Println("Get by id courses service")

	usuario, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return usuario, nil
}

func (s service) Delete(ctx context.Context, id string) error {
	s.log.Println("Delete by id courses service")

	err := s.repo.Delete(ctx, id)
	return err
}

func (s service) Update(ctx context.Context, id string, name *string, startDate, endDate *string) error {
	s.log.Println("Update course service")
	var startDateParsed *time.Time //Se crea un puntero *time.Time, ESTOS se crean en NIL. Si startDate y/o endDate tienen datos ENTRAN EN LOS IF de abajo y el puntero agarra un valor y direccion. SI NO QUEDAN EN NIL
	var endDateParsed *time.Time
	var err error

	if startDate != nil { //Si startDate viene nil es porque no veiene en el request, x lo que NO entra en el if
		parsedTime, err := time.Parse("2006-01-02", *startDate)
		if err != nil {
			s.log.Println(err)
			return err
		}
		startDateParsed = &parsedTime
	}
	if endDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *endDate)
		if err != nil {
			s.log.Println(err)
			return err
		}
		endDateParsed = &parsedTime
	}
	err = s.repo.Update(ctx, id, name, startDateParsed, endDateParsed)

	return err
}

func (s service) Count(ctx context.Context, filtros Filtros) (int, error) {
	s.log.Println("Count courses service")
	return s.repo.Count(ctx, filtros)
}
