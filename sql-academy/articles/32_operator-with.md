# Обобщённое табличное выражение, оператор WITH

Обобщённое табличное выражение или CTE (Common Table Expressions) - это временный результирующий набор данных, к которому можно обращаться в последующих запросах.
Для написания обобщённого табличного выражения используется оператор WITH.

```
MySQL 8.1
```

```
-- Пример использования конструкции WITH
WITH Aeroflot_trips AS
    (SELECT TRIP.* FROM Company
        INNER JOIN Trip ON Trip.company = Company.id WHERE name = 'Aeroflot')

SELECT plane, COUNT(plane) AS amount FROM Aeroflot_trips GROUP BY plane;

```

```

```

Выражение с WITH считается «временным», потому что результат не сохраняется где-либо на постоянной основе
в схеме базы данных, а действует как временное представление, которое существует только на время выполнения запроса,
то есть оно доступно только во время выполнения операторов SELECT, INSERT, UPDATE, DELETE или MERGE.
Оно действительно только в том запросе, которому он принадлежит, что позволяет улучшить структуру запроса, не загрязняя глобальное пространство имён.

## Синтаксис оператора WITH

```
MySQL 8.1
```

```
WITH название_cte [(столбец_1 [, столбец_2 ] …)] AS (подзапрос)
    [, название_cte [(столбец_1 [, столбец_2 ] …)] AS (подзапрос)] …

```

```

```

Порядок использования оператора WITH:

- Ввести оператор WITH

- Указать название обобщённого табличного выражения

- Опционально: определить названия для столбцов получившегося табличного выражения, разделённых знаком запятой

- Ввести AS и далее подзапрос, результат которого можно будет использовать в других частях SQL запроса, используя имя, определённое на 2 этапе

- Опционально: если необходимо более одного табличного выражения, то ставится запятая и повторяются шаги 2-4

## Примеры запросов

- Создаём табличное выражение Aeroflot_trips, содержащее все полёты, совершенные авиакомпанией «Aeroflot»

```
MySQL 8.1
```

```
WITH Aeroflot_trips AS
    (SELECT plane, town_from, town_to FROM Company
        INNER JOIN Trip ON Trip.company = Company.id WHERE name = 'Aeroflot')

SELECT * FROM Aeroflot_trips;

```

```

```

planetown_fromtown_toIL-86MoscowRostovIL-86RostovMoscow

- Аналогично, создаём табличное выражение Aeroflot_trips, но с переименованными колонками

```
MySQL 8.1
```

```
WITH Aeroflot_trips (aeroflot_plane, town_from, town_to) AS
    (SELECT plane, town_from, town_to FROM Company
        INNER JOIN Trip ON Trip.company = Company.id WHERE name = 'Aeroflot')

SELECT * FROM Aeroflot_trips;

```

```

```

aeroflot_planetown_fromtown_toIL-86MoscowRostovIL-86RostovMoscow

- С помощью оператора WITH определяем несколько табличных выражений

```
MySQL 8.1
```

```
WITH Aeroflot_trips AS
    (SELECT TRIP.* FROM Company
        INNER JOIN Trip ON Trip.company = Company.id WHERE name = 'Aeroflot'),
    Don_avia_trips AS
    (SELECT TRIP.* FROM Company
        INNER JOIN Trip ON Trip.company = Company.id WHERE name = 'Don_avia')

SELECT * FROM Don_avia_trips UNION SELECT * FROM  Aeroflot_trips;

```

```

```

idcompanyplanetown_fromtown_totime_outtime_in11811TU-134RostovMoscow1900-01-01T06:12:00.000Z1900-01-01T08:01:00.000Z11821TU-134MoscowRostov1900-01-01T12:35:00.000Z1900-01-01T14:30:00.000Z11871TU-134RostovMoscow1900-01-01T15:42:00.000Z1900-01-01T17:39:00.000Z11881TU-134MoscowRostov1900-01-01T22:50:00.000Z1900-01-02T00:48:00.000Z11951TU-154RostovMoscow1900-01-01T23:30:00.000Z1900-01-02T01:11:00.000Z11961TU-154MoscowRostov1900-01-01T04:00:00.000Z1900-01-01T05:45:00.000Z11452IL-86MoscowRostov1900-01-01T09:35:00.000Z1900-01-01T11:23:00.000Z11462IL-86RostovMoscow1900-01-01T17:55:00.000Z1900-01-01T20:01:00.000Z

## Работа с рекурсией в CTE

CTE также могут быть использованы для выполнения рекурсивных запросов,
которые позволяют итеративно обрабатывать данные, например,
для работы с иерархическими структурами данных, такими как «руководитель — подчинённый».

### Синтаксис рекурсивного CTE

Рекурсивное CTE состоит из двух частей, разделенных оператором UNION ALL:

- Начальный набор данных, который не содержит рекурсивных ссылок.

- Рекурсивная часть: запрос, который ссылается на CTE, чтобы продолжить рекурсию.

```
MySQL 8.1
```

```
WITH RECURSIVE название_cte (столбец_1, столбец_2, ...) AS (
    -- Начальный набор данных
    SELECT столбец_1, столбец_2, ...
    FROM таблица
    WHERE условие

    UNION ALL

    -- Рекурсивная часть
    SELECT столбец_1, столбец_2, ...
    FROM название_cte
    INNER JOIN таблица ON название_cte.столбец = таблица.столбец
    WHERE условие
)

SELECT * FROM название_cte;

```

```

```

### Пример: иерархия руководителей и подчинённых

Рассмотрим таблицу Employees, которая содержит идентификаторы сотрудников и их руководителей:

idnamemanagerId1John Smith<NULL>2Michael Johnson13Robert Williams14James Brown25David Jones26Richard Davis3
Требуется найти всех подчинённых John Smith (id=1) на всех уровнях иерархии.

```
MySQL 8.1
```

```
WITH RECURSIVE Subordinates AS (
    -- Начальный набор данных
    SELECT id, name, managerId
    FROM Employees
    WHERE managerId = 1

    UNION ALL

    -- Рекурсивная часть: подчинённые подчинённых
    SELECT e.id, e.name, e.managerId
    FROM Employees e
    INNER JOIN Subordinates s ON e.managerId = s.id
)

SELECT * FROM Subordinates;

```

```

```

idnamemanagerId2Michael Johnson13Robert Williams14James Brown25David Jones26Richard Davis3

### Шаги выполнения рекурсивного CTE

- Начальный набор данных: выбираются все сотрудники, у которых managerId=1 (непосредственные подчинённые John Smith).

- Рекурсивная часть: для каждого сотрудника, выбранного в начальном наборе данных, выбираются их подчинённые (где managerId равен id выбранного сотрудника).

- Объединение: результаты начального набора данных и рекурсивной частей объединяются с помощью UNION ALL.

- Рекурсия: процесс повторяется для каждого нового набора подчинённых, пока не будут выбраны все уровни иерархии.

## Заключение

Обобщённые табличные выражения были добавлены в SQL для упрощения сложных длинных запросов, особенно с множественными подзапросами. Их главная задача – улучшение читабельности,
простоты написания запросов и их дальнейшей поддержки. Это происходит за счёт сокрытия больших и сложных запросов в созданные именованные выражения, которые потом используются в основном запросе.