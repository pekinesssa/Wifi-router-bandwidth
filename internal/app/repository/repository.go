package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
  return &Repository{}, nil
}


type Order struct { // вот наша новая структура 
  ID    int // поля структур, которые передаются в шаблон
  Title string // ОБЯЗАТЕЛЬНО должны быть написаны с заглавной буквы (то есть публичными)
  Description string
  Price string
  Image string
}

type Cart struct { // вот наша новая структура 
  ID    int // поля структур, которые передаются в шаблон
  Goods []int // ОБЯЗАТЕЛЬНО должны быть написаны с заглавной буквы (то есть публичными)
  Address string
  CountOrders int
  Formula string
}

func (r *Repository) GetOrders() ([]Order, error) {
  // имитируем работу с БД. Типа мы выполнили sql запрос и получили эти строки из БД
  orders := []Order{ // массив элементов из наших структур
    {
      ID:    1,
      Title: "Базовый план",
	  Description: "Стандартный пакет для семьи.",
	  Price: "50000.00 руб",
    Image:  "http://localhost:9000/main/1.png",
    },
    {
      ID:    2,
      Title: "Стриминг и развлечения",
	  Description: "Идеально для долгих сериалов и фильмов.",
	  Price: "65000.00 руб",
    Image:  "http://localhost:9000/main/2.png",

    },
    {
      ID:    3,
      Title: "Онлайн-игры и стриминг игр",
	  Description: "Низкий пинг высокая отдача скорости.",
	  Price: "85000.00 руб",
    Image:  "http://localhost:9000/main/3.png",

    },
  }
  // обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
  // тут я снова искусственно обработаю "ошибку" чисто чтобы показать вам как их передавать выше
  if len(orders) == 0 {
    return nil, fmt.Errorf("массив пустой")
  }

  return orders, nil
}

func (r *Repository) GetCart(id int) (Cart, error) {
	if id != 1 {
		return Cart{}, fmt.Errorf("корзина с id %d не найдена", id)
	}

	allOrders, err := r.GetOrders()
	if err != nil {
		return Cart{}, fmt.Errorf("не удалось получить список товаров для создания корзины: %w", err)
	}

	goodsIDs := []int{}
	firstOrder := allOrders[0]
	goodsIDs = append(goodsIDs, firstOrder.ID)

	secondOrder := allOrders[1]
	goodsIDs = append(goodsIDs, secondOrder.ID)

	cart := Cart{
		ID:          1,
		Goods:       goodsIDs, 
		Address:     "ул. Проспект мира, д. 5 с.7, кв. 579",
		CountOrders: len(goodsIDs), 
		Formula:     "500 Мбит/с",
	}

	return cart, nil
}

func (r *Repository) GetOrder(id int) (Order, error) {
	// тут у вас будет логика получения нужной услуги, тоже наверное через цикл в первой лабе, и через запрос к БД начиная со второй 
	orders, err := r.GetOrders()
	if err != nil {
		return Order{}, err // тут у нас уже есть кастомная ошибка из нашего метода, поэтому мы можем просто вернуть ее
	}

	for _, order := range orders {
		if order.ID == id {
			return order, nil // если нашли, то просто возвращаем найденный заказ (услугу) без ошибок
		}
	}
	return Order{}, fmt.Errorf("заказ не найден") // тут нужна кастомная ошибка, чтобы понимать на каком этапе возникла ошибка и что произошло
}

func (r *Repository) GetOrdersByTitle(title string) ([]Order, error) {
	orders, err := r.GetOrders()
	if err != nil {
		return []Order{}, err
	}

	var result []Order
	for _, order := range orders {
		if strings.Contains(strings.ToLower(order.Title), strings.ToLower(title)) {
			result = append(result, order)
		}
	}

	return result, nil
}