# Операторы IS NULL, BETWEEN, IN

Мы уже познакомились с синтаксисом оператора WHERE и операторами сравнения, но помимо них в условных запросах мы можем использовать следующие полезные операторы:

- IS NULL

- BETWEEN

- IN

Давайте рассмотрим их применение.

## IS NULL

Оператор IS NULL позволяет проверить, отсутствует ли значение в поле.
Для примера выведем всех преподавателей, у кого отсутствует отчество:

```
MySQL 8.1
```

```
SELECT * FROM Teacher
WHERE middle_name IS NULL;

```

```

```

idfirst_namemiddle_namelast_name10YUrij<NULL>Krylov11Andrej<NULL>Evseev
Для использования отрицания, то есть, если мы хотим найти все записи, где поле не равно NULL, мы должны использовать следующий синтаксис:

```
MySQL 8.1
```

```
SELECT * FROM Teacher
WHERE middle_name IS NOT NULL;

```

```

```

## BETWEEN

Оператор BETWEEN min AND max позволяет узнать, расположено ли проверяемое значение столбца в интервале между min и max, включая сами значения min и max.
Он идентичен условию:

```
MySQL 8.1
```

```
... WHERE field >= min AND field <= max

```

```

```

Используется данный оператор следующим образом:

```
MySQL 8.1
```

```
SELECT * FROM Payments
WHERE unit_price BETWEEN 100 AND 500;

```

```

```

В качестве результата вернутся все записи из таблицы Payments, где значение поля unit_price будет от 100 до 500.

## IN

Оператор IN позволяет узнать, входит ли проверяемое значение столбца в список определённых значений.

```
MySQL 8.1
```

```
SELECT * FROM FamilyMembers
WHERE status IN ('father', 'mother');

```

```

```

member_idstatusmember_namebirthday1fatherHeadley Quincey1960-05-13T00:00:00.000Z2motherFlavia Quincey1963-02-16T00:00:00.000Z6fatherErnest Forrest1961-09-11T00:00:00.000Z7motherConstance Forrest1968-09-06T00:00:00.000Z