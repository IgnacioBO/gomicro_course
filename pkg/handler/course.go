package handler

//AQUI ESTARAN LOS RUTEOS Y LOS MIDDLWARE

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/IgnacioBO/go_lib_response/response"
	"github.com/IgnacioBO/gomicro_course/internal/course"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

// Este recibe contexto y un endpoint que definimos en la capa del endpionts
func NewUserHTTPServer(ctx context.Context, endpoints course.Endpoints) http.Handler {
	//GIN (con mux era router := mux.NewRouter())
	router := gin.Default()

	//Esta se guarad en opciones y se pone al final en Handle
	opciones := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
	}

	//Ahora usaremos Handle, poreque a este se le puede pasar un server (httptranpsort)
	//Ahora usaremos GIN en vez de mux
	//router.POST recibe un string y una funcion [HandlerFunc] (antes aceptaba una interface http.Handler)
	//Entonces tenemos un WrhapH que toma este http.Handler y devuelve un HandlerFunc para que quede ok y router.POST no de error
	router.POST("/courses", gin.WrapH(httptransport.NewServer(
		endpoint.Endpoint(endpoints.Create),
		decodeCreateCourse,
		encodeResponse,
		opciones...,
	)))

	/* Post antiguo con Mux
	router.Handle("/courses", httptransport.NewServer(
		endpoint.Endpoint(endpoints.Create),
		decodeCreateCourse,
		encodeResponse,
		opciones...,
	)).Methods("POST")*/

	router.GET("/courses", gin.WrapH(httptransport.NewServer(
		endpoint.Endpoint(endpoints.GetAll),
		decodeGetAllCourse,
		encodeResponse,
		opciones...,
	)))

	//Aparte del gin.WrapH; le enviaremos ANTES un ginDecode (que es un metodo que creamos que hace que el request tenga un contexto con "params" que son los parametros)
	router.GET("/courses/:id", ginDecode, gin.WrapH(httptransport.NewServer(
		endpoint.Endpoint(endpoints.Get),
		decodeGetCourse,
		encodeResponse,
		opciones...,
	)))

	/* Get antiguo con Mux
	router.Handle("/courses/{id}", httptransport.NewServer(
		endpoint.Endpoint(endpoints.Get),
		decodeGetCourse,
		encodeResponse,
		opciones...,
	)).Methods("GET")*/

	router.DELETE("/courses/:id", ginDecode, gin.WrapH(httptransport.NewServer(
		endpoint.Endpoint(endpoints.Delete),
		decodeDeleteCourse,
		encodeResponse,
		opciones...,
	)))

	router.PATCH("/courses/:id", ginDecode, gin.WrapH(httptransport.NewServer(
		endpoint.Endpoint(endpoints.Update),
		decodeUpdateCourse,
		encodeResponse,
		opciones...,
	)))

	return router
}

// MIDDLEWARE DE GIN, QUE RCIEB UN CONTEXTO DE GIN, OSEA RECIBE UN CONTEXTO DE GIN gin.Context y TOMA EL CONTEXTO de GIN
// Entonces le pasaremos por contexto, definiermos un valor en el contexto q se llamara "params" que tendra los aprametros
// Entonces ahora el contexto de tipo context.Context [que es el que usa nuestros decodes] (NO el gin.Context) tendra el contexto de gin.Context!
func ginDecode(c *gin.Context) {
	ctx := context.WithValue(c.Request.Context(), "params", c.Params) //Le decimos que que el ctx tendra una llave "params" que devueve un c.Params [que sirve para obtener parametros]
	c.Request = c.Request.WithContext(ctx)                            //Genero el request con el contexto, para que el requeste VALLA CON ESE CONTEXT y pueda obtenerse en los decode
}

// *** MIDDLEWARE REQUEST ***
func decodeCreateCourse(_ context.Context, r *http.Request) (interface{}, error) {
	//Le pasamos un validador de ahtotization (que le pasamos del token del header)
	//Si es invalida da error de autorizacion
	if err := authorization(r.Header.Get("Authorization")); err != nil {
		return nil, response.Forbidden(err.Error())
	}

	var reqStruct course.CreateRequest

	//Ahora hacemos el decode del body del json al srtuct de REquest de course
	err := json.NewDecoder(r.Body).Decode(&reqStruct)
	if err != nil {
		return nil, response.BadRequest(fmt.Sprintf("invalid request format: '%v'", err.Error()))
	}
	return reqStruct, nil
}

// *** MIDDLEWARE RESPONSE ***
func encodeResponse(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	rInterface := resp.(response.Response)                            //Transformamos el resp a response.Respone (al interface) -> YA QUE LE ENAIREMOS SIEMPRE UN objeto RESPONSE (CREADO POR NOSOTROS, q tiene el code, mensage, meta, etc, todo el json)
	w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json
	w.WriteHeader(rInterface.StatusCode())
	return json.NewEncoder(w).Encode(rInterface) //resp tendra el user.User del domain y otroas datos si es necesario para ocnveritse en json

}

// *** MIDDLEWARE RESPONSE DE ERROR ***
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json
	respInterface := err.(response.Response)                          //Tranfosrmamos el error recibido a la interfac response.Response que craemos
	//¿Porque funciona esta conversion de tipo error al de nosotros?, porque la interfaz 'error' de go pide que haya un metodo Error() string [QUE CREAMOS EN nuestro respon.RESPONSE!]
	//Entonces como implementamos el metodo Error() string funcinoa, ademas tenemos al ventaja que vamos apoder obtener MAS DATOS porque repsonse.Response tiene mas metodos como (StatusCode())
	//Entonces podemos transofrmar un error a una interfac propia con MAS METODOS Y MAS DATOS UE UN ERROR NORMAL!
	w.WriteHeader(respInterface.StatusCode())
	_ = json.NewEncoder(w).Encode(respInterface)

}

// *** MIDDLEWARE REQUEST GET ***
// Ahora SI RECIBIERMO EL CONTEXTO, con ctx (antes _)
func decodeGetCourse(ctx context.Context, r *http.Request) (interface{}, error) {
	//Le pasamos un validador de ahtotization (que le pasamos del token del header)
	if err := authorization(r.Header.Get("Authorization")); err != nil {
		return nil, response.Forbidden(err.Error())
	}

	var getReq course.GetRequest
	/* Antiguio metodo con mux
	variablesPath := mux.Vars(r)
	getReq.ID = variablesPath["id"] //OBtenemos el id y lo guardamos en el cmapo ID de getReq
	*/

	//Obtendremos params, que se lo dimos con la funcion ginDecode
	//Value devuelve un any/interface [osea cualquier cosa], tenemos que convertir al objeto que quiero recibir (el parametro de gin, gin.Params)
	params := ctx.Value("params").(gin.Params) // .(gin.Params) es para transfmar o asegur que el objeto sera un gun.Params

	getReq.ID = params.ByName("id") //Con byName obtenemos el valor de id

	fmt.Println("id es:", getReq.ID)

	return getReq, nil

}

// *** MIDDLEWARE REQUEST GET All ***
// Funcion de decode, de GET
func decodeGetAllCourse(_ context.Context, r *http.Request) (interface{}, error) {
	//Le pasamos un validador de ahtotization (que le pasamos del token del header)
	if err := authorization(r.Header.Get("Authorization")); err != nil {
		return nil, response.Forbidden(err.Error())
	}

	//Query() devielve un objeto que permite acceder a los parametros d la url (...?campo=123&campo2=hola)
	variablesURL := r.URL.Query()

	//Ahora obtendremos el limit y la pagina desde los parametros
	limit, _ := strconv.Atoi(variablesURL.Get("limit"))
	page, _ := strconv.Atoi(variablesURL.Get("page"))

	getReqAll := course.GetAllRequest{
		Name:  variablesURL.Get("name"),
		Limit: limit,
		Page:  page,
	}

	return getReqAll, nil
}

// *** MIDDLEWARE REQUEST Delete ***
func decodeDeleteCourse(ctx context.Context, r *http.Request) (interface{}, error) {
	//Le pasamos un validador de ahtotization (que le pasamos del token del header)
	if err := authorization(r.Header.Get("Authorization")); err != nil {
		return nil, response.Forbidden(err.Error())
	}

	params := ctx.Value("params").(gin.Params) // .(gin.Params) es para transfmar o asegur que el objeto sera un gun.Params
	id := params.ByName("id")                  //Con byName obtenemos el valor de id

	fmt.Println("id a eliminar es:", id)
	deleteReq := course.DeleteRequest{ID: id}

	return deleteReq, nil

}

// *** MIDDLEWARE REQUEST Delete***
func decodeUpdateCourse(ctx context.Context, r *http.Request) (interface{}, error) {
	//Le pasamos un validador de ahtotization (que le pasamos del token del header)
	if err := authorization(r.Header.Get("Authorization")); err != nil {
		return nil, response.Forbidden(err.Error())
	}

	var reqStruct course.UpdateRequest

	err := json.NewDecoder(r.Body).Decode(&reqStruct)
	if err != nil {
		return nil, response.BadRequest(fmt.Sprintf("invalid request format: '%v'", err.Error()))
	}
	print("Lol")

	params := ctx.Value("params").(gin.Params) // .(gin.Params) es para transfmar o asegur que el objeto sera un gun.Params
	reqStruct.ID = params.ByName("id")         //Con byName obtenemos el valor de id

	return reqStruct, nil

}

// Authoruzation con otken
func authorization(token string) error {
	if token != os.Getenv("TOKEN") {
		return errors.New("invalid token")
	}
	return nil
}
