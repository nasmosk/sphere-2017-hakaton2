- завести нового пользователя
    example: "/newuser?u_id=123&cash=1000000"
    params:
        u_id - int
        cash - int 
    response:
    {
        "response":{
            "id":123123 //id в базе
        }
    }

- посмотреть свои позиции и баланс
    example: "/getinfo?u_id=123"
    params:
        u_id - int
    response:
    {
        "response":{
            "balance":100000,
            "records":[
                {
                    "id": 111,
                    "ticker": "SPFB.BR-1.18",
                    "vol"   : 12
                },
                {
                    "id": 111,
                    "ticker": "SPFB.BR-1.18",
                    "vol"   : 10
                }
            ]
        }
    }

- посмотрить историю своих сделок
    example: "/gethistory?u_id=123"
    params:
        u_id - int
    response:
    {
        "response":{
            "records":[
                {
                    "id": 111,
                    "ticker": "SPFB.BR-1.18",
                    "is_buy": 0,
                    "price" : 65,
                    "vol"   : 12
                },
                {
                    "id": 111,
                    "ticker": "SPFB.BR-1.18",
                    "is_buy": 1,
                    "price" : 67,
                    "vol"   : 10
                }
            ]
        }
    }

- отправить на биржу заявку на покупку тикера
    example: "/buyreq?u_id=123&ticker=SPFB.BR-1.18&vol=123&price=65"
    params:
        u_id - int 
        ticker - str
        vol - int
        price - int 
    response:
    {
        "response":{
            "id":321 //id заявки
        }
    }

- отправить на биржу заявку на продажу тикера
    example: "/sellreq?u_id=123&ticker=SPFB.BR-1.18&vol=123&price=65"
    params:
        u_id - int 
        ticker - str
        vol - int
        price - int 
    response:
    {
        "response":{
            "id":321 //id заявки
        }
    }

- отменить ранее отправленную заявку
    example: "/abortreq?u_id=123&req_id=228"
    params:
        u_id - int 
        req_id - int
    response:
    {
        "response":{
            "id":321 //id отмененной заявки
        }
    }

- посмотреть последнюю истории торгов
    example: "/tradehistory"
    resp: //TODO


Error response:
    {
        "error":"wrong method"
    }