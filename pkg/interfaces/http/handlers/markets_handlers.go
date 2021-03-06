package handlers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/ralvescosta/base/pkg/app/interfaces"
	"github.com/ralvescosta/base/pkg/domain/usecases"
	httpServer "github.com/ralvescosta/base/pkg/infra/http_server"
	"github.com/ralvescosta/base/pkg/interfaces/http/factories"
	viewmodels "github.com/ralvescosta/base/pkg/interfaces/http/view_models"
)

type IMarketHandlers interface {
	Create(httpRequest httpServer.HttpRequest) httpServer.HttpResponse
	GetByQuery(httpRequest httpServer.HttpRequest) httpServer.HttpResponse
	Update(httpRequest httpServer.HttpRequest) httpServer.HttpResponse
	Delete(httpRequest httpServer.HttpRequest) httpServer.HttpResponse
}

type marketHandlers struct {
	logger              interfaces.ILogger
	validator           interfaces.IValidator
	httpResFactory      factories.HttpResponseFactory
	createUseCase       usecases.ICreateMarketUseCase
	getByQueryUseCase   usecases.IGetMarketByQueryUseCase
	updateMarketUseCase usecases.IUpdateMarketUseCase
	deleteUseCase       usecases.IDeleteMarketUseCase
}

func (pst marketHandlers) Create(httpRequest httpServer.HttpRequest) httpServer.HttpResponse {
	vModel := viewmodels.MarketViewModel{}
	if err := json.Unmarshal(httpRequest.Body, &vModel); err != nil {
		return pst.httpResFactory.BadRequest("body is required", nil)
	}

	if validationErrs := pst.validator.ValidateStruct(vModel); validationErrs != nil {
		pst.logger.Error(fmt.Sprintf("[MarketHandler::Create] - Body unformatted - %s", validationErrs[0].Message))
		return pst.httpResFactory.BadRequest(validationErrs[0].Message, nil)
	}

	result, alreadyCreated, err := pst.createUseCase.Execute(httpRequest.Ctx, vModel.ToValueObject())
	if err != nil {
		return pst.httpResFactory.ErrorResponseMapper(err, nil)
	}
	if alreadyCreated {
		return pst.httpResFactory.Ok(viewmodels.NewMarketViewModel(result), nil)
	}

	return pst.httpResFactory.Created(viewmodels.NewMarketViewModel(result), nil)
}

func (pst marketHandlers) GetByQuery(httpRequest httpServer.HttpRequest) httpServer.HttpResponse {
	vModel, err := queryToMarketViewModel(httpRequest.Query)
	if err != nil {
		return pst.httpResFactory.BadRequest(err.Error(), nil)
	}

	result, err := pst.getByQueryUseCase.Execute(httpRequest.Ctx, vModel.ToValueObject())
	if err != nil {
		return pst.httpResFactory.ErrorResponseMapper(err, nil)
	}

	return pst.httpResFactory.Ok(viewmodels.NewSliceOfMarketViewModel(result), nil)
}

func queryToMarketViewModel(query map[string][]string) (viewmodels.MarketViewModel, error) {
	vModel := viewmodels.MarketViewModel{}
	voReflect := reflect.ValueOf(&vModel)
	for k, v := range query {
		var ff reflect.Value
		if k == "nome_feira" {
			ff = voReflect.Elem().FieldByName("NomeFeira")
		} else {
			ff = voReflect.Elem().FieldByName(strings.Title(k))
		}

		if ff.Kind() == 0 {
			return viewmodels.MarketViewModel{}, fmt.Errorf("paramter: %s not allowed", k)
		}

		if ff.Type().Name() == "int" {
			t, err := strconv.ParseInt(v[0], 10, 64)
			if err != nil {
				return viewmodels.MarketViewModel{}, fmt.Errorf("paramter: %s is not a valid integer", k)
			}
			ff.SetInt(t)
		} else {
			ff.SetString(v[0])
		}
	}

	return vModel, nil
}

func (pst marketHandlers) Update(httpRequest httpServer.HttpRequest) httpServer.HttpResponse {
	vModel := viewmodels.MarketViewModel{}
	if err := json.Unmarshal(httpRequest.Body, &vModel); err != nil {
		return pst.httpResFactory.BadRequest("body is required", nil)
	}
	if vModel.Registro != "" {
		return pst.httpResFactory.BadRequest("the field 'registro' is not allowed", nil)
	}

	registerCode, ok := httpRequest.Params["registerCode"]
	if !ok {
		return pst.httpResFactory.BadRequest("registerCode is required", nil)
	}

	result, err := pst.updateMarketUseCase.Execute(httpRequest.Ctx, registerCode, vModel.ToValueObject())
	if err != nil {
		return pst.httpResFactory.ErrorResponseMapper(err, nil)
	}

	return pst.httpResFactory.Ok(viewmodels.NewMarketViewModel(result), nil)
}

func (pst marketHandlers) Delete(httpRequest httpServer.HttpRequest) httpServer.HttpResponse {
	registerCode, ok := httpRequest.Params["registerCode"]
	if !ok {
		return pst.httpResFactory.BadRequest("registerCode is required", nil)
	}

	if err := pst.deleteUseCase.Execute(httpRequest.Ctx, registerCode); err != nil {
		return pst.httpResFactory.ErrorResponseMapper(err, nil)
	}

	return pst.httpResFactory.Ok(struct{}{}, nil)
}

func NewMarketHandlers(logger interfaces.ILogger, validator interfaces.IValidator, httpResFactory factories.HttpResponseFactory,
	createUseCase usecases.ICreateMarketUseCase, getByQueyUseCase usecases.IGetMarketByQueryUseCase, updateMarketUseCase usecases.IUpdateMarketUseCase,
	deleteUseCase usecases.IDeleteMarketUseCase) IMarketHandlers {

	return marketHandlers{
		logger,
		validator,
		httpResFactory,
		createUseCase,
		getByQueyUseCase,
		updateMarketUseCase,
		deleteUseCase,
	}
}
