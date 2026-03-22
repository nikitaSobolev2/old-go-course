package domain // доменные ошибки-отправители для маппинга в HTTP

import "errors"

var ErrNotFound = errors.New("not found") // заказ с таким id не найден в репозитории
