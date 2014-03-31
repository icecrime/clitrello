package main

import (
	"fmt"
	"net/http"
)

type Action interface {
	BuildURL(*Config) string
	HandleResponse(*http.Response) []interface{}
}

func Execute(config *Config, action Action) (result []interface{}) {
	endpoint := action.BuildURL(config)

	resp, err := http.Get(endpoint)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode == 200 {
		result = action.HandleResponse(resp)
	} else {
		panic(fmt.Sprintf("Call failed (error code %d)", resp.StatusCode))
	}

	return
}

/**
 ******************************************************************************
 */

type ListBoards struct{}

func NewListBoardsAction() *ListBoards {
	return &ListBoards{}
}

func (*ListBoards) BuildURL(config *Config) string {
	return TrelloURL(config, "members/me/boards", nil)
}

func (*ListBoards) HandleResponse(response *http.Response) []interface{} {
	return GetJSONContent(response)
}

/**
 ******************************************************************************
 */

type GetBoardCards struct {
	boardId string
}

func NewGetBoardCardsAction(boardId string) *GetBoardCards {
	return &GetBoardCards{boardId}
}

func (action *GetBoardCards) BuildURL(config *Config) string {
	path := "boards/" + action.boardId + "/lists"
	params := map[string]string{"cards": "all"}
	return TrelloURL(config, path, params)
}

func (*GetBoardCards) HandleResponse(response *http.Response) []interface{} {
	return GetJSONContent(response)
}

/**
 ******************************************************************************
 */

func AuthorizationURL(config *Config, appName string) string {
	params := map[string]string{
		"key":          config.ApiKey,
		"name":         appName,
		"reponse_type": "token",
		"scope":        "read,write",
	}

	// We don't rely on the TrelloURL utility function because this is the one
	// case where we shouldn't provide a user token even if we do have one.
	endpoint := "https://trello.com/1/authorize"
	return BuildURL(endpoint, params)
}
