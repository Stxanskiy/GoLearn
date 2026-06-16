# Партиции в оконных функциях

В прошлой статье мы кратко уже упоминали, что
такое партиции и как их использовать в оконных функциях, пришло время разобраться в них поподробнее 🤓.

## Понятие партиции

Партиции — подмножества строк, выделенные для оконной функции на основе одного или нескольких столбцов в таблице.

Они служат для сегментации данных, позволяя выполнить более детальный анализ и
расчёты вроде агрегации или ранжирования внутри каждой такой группы.
Применяя партиционирование, например, по типу жилья в таблице с данными о цене жилья,
мы можем рассчитать в отдельной колонке, скажем, среднюю цену для каждого типа жилья.

## Применение партиций в SQL

Для того чтобы использовать партицию вместе с оконной функцией, необходимо придерживаться следующего
синтаксиса:

```
MySQL 8.1
```

```
SELECT <оконная_функция>(<поле_таблицы>)
OVER (
    PARTITION BY <столбцы_для_разделения>
)

```

```

```

### Пример использования

А теперь давайте на простом примере рассмотрим использование партиции вместе с оконной функцией.

Для этого рассмотрим таблицу Rooms, а именно поля home_type и price:

```
MySQL 8.1
```

```
SELECT home_type, price FROM Rooms;

```

```

```

home_typepricePrivate room149Entire home/apt225Private room150Entire home/apt89Entire home/apt80Entire home/apt200Private room60Private room79Private room79Entire home/apt150Entire home/apt135Private room85Private room89Private room85Entire home/apt120Entire home/apt140Entire home/apt215Private room140Entire home/apt99Entire home/apt190Entire home/apt299Private room130Private room80Private room110Entire home/apt120Private room60Private room80Entire home/apt150Private room44Entire home/apt180Private room50Private room52Private room55Private room50Private room70Private room89Private room35Entire home/apt85Private room150Shared room40Private room68Entire home/apt120Private room120Private room135Entire home/apt150Entire home/apt150Private room130Entire home/apt110Entire home/apt115Private room80
Мы можем увидеть, что все жильё для аренды разделено на 3 категории: «Private room», «Entire home/apt» и «Shared room».
Каждая категория жилья имеет свои ценовые рамки.
Чтобы узнать среднюю стоимость в конкретной категории и сравнить её с текущей, как раз можно использовать оконные функции.
Для этого добавим к нашей результирующей таблице ещё одно поле avg_price, которое будет высчитывать среднюю цену по категориям. Это будет выглядеть следующим образом:

```
MySQL 8.1
```

```
SELECT
    home_type, price,
    AVG(price) OVER (PARTITION BY home_type) AS avg_price
FROM Rooms

```

```

```

home_typepriceavg_priceEntire home/apt225148.6667Entire home/apt180148.6667Entire home/apt150148.6667Entire home/apt85148.6667Entire home/apt120148.6667Entire home/apt120148.6667Entire home/apt299148.6667Entire home/apt190148.6667Entire home/apt99148.6667Entire home/apt215148.6667Entire home/apt140148.6667Entire home/apt120148.6667Entire home/apt150148.6667Entire home/apt135148.6667Entire home/apt150148.6667Entire home/apt110148.6667Entire home/apt115148.6667Entire home/apt200148.6667Entire home/apt150148.6667Entire home/apt80148.6667Entire home/apt89148.6667Private room6889.4286Private room5089.4286Private room7089.4286Private room8089.4286Private room8989.4286Private room14989.4286Private room3589.4286Private room15089.4286Private room13089.4286Private room12089.4286Private room13589.4286Private room13089.4286Private room15089.4286Private room6089.4286Private room7989.4286Private room7989.4286Private room8589.4286Private room8989.4286Private room8589.4286Private room14089.4286Private room5589.4286Private room8089.4286Private room11089.4286Private room6089.4286Private room8089.4286Private room4489.4286Private room5089.4286Private room5289.4286Shared room4040
Что именно происходит в добавленной строке?

- PARTITION BY home_type делит все записи на разные партиции на основе уникальных значений столбца home_type

- затем для каждой записи AVG(price) вычисляет среднюю цену (price) в рамках её партиции (home_type)

Результатом выполнения этой части запроса будет столбец avg_price,
в котором для каждой записи будет указано среднее значение цены для её категории жилья (home_type).

## Партиции по нескольким колонкам

Партиционирование также может быть выполнено по нескольким колонкам. Это позволяет создавать более сложные и точные сегменты для анализа.
Например, для нашей таблицы Rooms мы можем создать партиции на основании 2 колонок: категория жилья
home_type и наличие телевизора в жилье has_tv.
Пример запроса с партиционированием по двум столбцам:

```
MySQL 8.1
```

```
SELECT
    home_type, has_tv, price,
    AVG(price) OVER (PARTITION BY home_type, has_tv) AS avg_price
    FROM Rooms

```

```

```

home_typehas_tvpriceavg_priceEntire home/apt0225170Entire home/apt0180170Entire home/apt080170Entire home/apt0200170Entire home/apt0150170Entire home/apt0150170Entire home/apt0190170Entire home/apt0215170Entire home/apt0140170Entire home/apt199132.6667Entire home/apt185132.6667Entire home/apt1150132.6667Entire home/apt1120132.6667Entire home/apt1120132.6667Entire home/apt1299132.6667Entire home/apt1120132.6667Entire home/apt1135132.6667Entire home/apt1150132.6667Entire home/apt1110132.6667Entire home/apt189132.6667Entire home/apt1115132.6667Private room08578.5455Private room03578.5455Private room015078.5455Private room05578.5455Private room05278.5455Private room05078.5455Private room06878.5455Private room06078.5455Private room013578.5455Private room08578.5455Private room08978.5455Private room112096.4706Private room18096.4706Private room114996.4706Private room113096.4706Private room18996.4706Private room17096.4706Private room15096.4706Private room14496.4706Private room18096.4706Private room16096.4706Private room111096.4706Private room18096.4706Private room113096.4706Private room114096.4706Private room17996.4706Private room17996.4706Private room115096.4706Shared room14040
Здесь PARTITION BY home_type, has_tv создаёт уникальные партиции для каждой комбинации home_type и has_tv,
позволяя вычислить среднюю цену жилья для текущей категории жилья с наличием или без наличия телевизора.