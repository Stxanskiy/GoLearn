# Внутреннее соединение INNER JOIN

В предыдущем уроке мы рассмотрели общую структуру многотабличного запроса:

```
MySQL 8.1
```

```
SELECT поля_таблиц
FROM таблица_1
[INNER] | [[LEFT | RIGHT | FULL][OUTER]] JOIN таблица_2
    ON условие_соединения
[[INNER] | [[LEFT | RIGHT | FULL][OUTER]] JOIN таблица_n
    ON условие_соединения]

```

```

```

Говоря о многотабличном запросе со внутренним соединением, общая структура выглядит так:

```
MySQL 8.1
```

```
SELECT поля_таблиц
FROM таблица_1
[INNER] JOIN таблица_2
    ON условие_соединения
[[INNER] JOIN таблица_n
    ON условие_соединения]

```

```

```

Например, запрос может выглядеть следующим образом:

```
MySQL 8.1
```

```
SELECT family_member, member_name FROM Payments
INNER JOIN FamilyMembers
    ON Payments.family_member = FamilyMembers.member_id

```

```

```

family_membermember_name1Headley Quincey2Flavia Quincey3Andie Quincey4Lela Quincey4Lela Quincey5Annie Quincey2Flavia Quincey2Flavia Quincey5Annie Quincey3Andie Quincey2Flavia Quincey1Headley Quincey3Andie Quincey3Andie Quincey
Так как, по умолчанию, если не указаны какие-либо параметры, JOIN выполняется как INNER JOIN, то при внутреннем соединении INNER является опциональным.

## Понятие внутреннего соединения

Внутреннее соединение — соединение, при котором находятся пары записей из двух таблиц, удовлетворяющие условию соединения, тем самым образуя новую таблицу, содержащую
поля из первой и второй исходных таблиц.
Для наглядности это выглядит следующим образом:

Так как в нашем условии указано равенство полей Payments.good_id и Goods.good_id, то при внутреннем соединении в итоговой выборке окажутся только записи,
где в обеих таблицах есть одинаковое значение good_id.

## Использование WHERE для соединения таблиц

Для внутреннего соединения таблиц также можно использовать оператор WHERE. Например, вышеприведённый запрос, написанный с помощью INNER JOIN, будет выглядеть так:

```
MySQL 8.1
```

```
SELECT family_member, member_name FROM Payments, FamilyMembers
    WHERE Payments.family_member = FamilyMembers.member_id

```

```

```

family_membermember_name1Headley Quincey2Flavia Quincey3Andie Quincey4Lela Quincey4Lela Quincey5Annie Quincey2Flavia Quincey2Flavia Quincey5Annie Quincey3Andie Quincey2Flavia Quincey1Headley Quincey3Andie Quincey3Andie Quincey