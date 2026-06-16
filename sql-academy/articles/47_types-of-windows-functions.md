# Основные оконные функции

В предыдущих статьях мы рассмотрели, как работают оконные функции, познакомились с понятием окна данных,
которое передаётся в оконную функцию. Пришло время рассмотреть какие оконные функции бывают.

## Виды оконных функций

Оконные функции можно разделить на 3 группы:

- Агрегатные оконные функции

- Ранжирующие оконные функции

- Оконные функции смещения

### Агрегатные оконные функции

Агрегатные функции — это функции, которые выполняют на наборе данных арифметические вычисления и возвращают итоговое значение.

- SUM — подсчитывает общую сумму значений;

- COUNT — считает общее количество записей в колонке;

- AVG — рассчитывает среднее арифметическое;

- MAX — находит наибольшее значение;

- MIN — определяет наименьшее значение.

```
MySQL 8.1
```

```
SELECT id,
	home_type,
	price,
	SUM(price) OVER(PARTITION BY home_type) AS "Sum",
	COUNT(price) OVER(PARTITION BY home_type) AS "Count",
	AVG(price) OVER(PARTITION BY home_type) AS "Avg",
	MAX(price) OVER(PARTITION BY home_type) AS "Max",
	MIN(price) OVER(PARTITION BY home_type) AS "Min"
FROM Rooms;

```

```

```

idhome_typepriceSumCountAvgMaxMin2Entire home/apt225312221148.66672998030Entire home/apt180312221148.66672998028Entire home/apt150312221148.66672998038Entire home/apt85312221148.66672998025Entire home/apt120312221148.66672998042Entire home/apt120312221148.66672998021Entire home/apt299312221148.66672998020Entire home/apt190312221148.66672998019Entire home/apt99312221148.66672998017Entire home/apt215312221148.66672998016Entire home/apt140312221148.66672998015Entire home/apt120312221148.66672998046Entire home/apt150312221148.66672998011Entire home/apt135312221148.66672998010Entire home/apt150312221148.66672998048Entire home/apt110312221148.66672998049Entire home/apt115312221148.6667299806Entire home/apt200312221148.66672998045Entire home/apt150312221148.6667299805Entire home/apt80312221148.6667299804Entire home/apt89312221148.66672998041Private room6825042889.42861503534Private room5025042889.42861503535Private room7025042889.42861503550Private room8025042889.42861503536Private room8925042889.4286150351Private room14925042889.42861503537Private room3525042889.42861503539Private room15025042889.42861503547Private room13025042889.42861503543Private room12025042889.42861503544Private room13525042889.42861503522Private room13025042889.4286150353Private room15025042889.4286150357Private room6025042889.4286150358Private room7925042889.4286150359Private room7925042889.42861503512Private room8525042889.42861503513Private room8925042889.42861503514Private room8525042889.42861503518Private room14025042889.42861503518Private room14025042889.42861503533Private room5525042889.42861503523Private room8025042889.42861503524Private room11025042889.42861503526Private room6025042889.42861503527Private room8025042889.42861503529Private room4425042889.42861503531Private room5025042889.42861503532Private room5225042889.42861503540Shared room40401404040

### Ранжирующие оконные функции

Ранжирующие оконные функции — это функции, которые ранжируют значение для каждой строки в окне.
В ранжирующих функциях под ключевым словом OVER обязательным идёт указание условия ORDER BY, по которому будет происходить сортировка ранжирования.

- ROW_NUMBER — возвращает номер строки, используется для нумерации;

- RANK — возвращает ранг каждой строки. Вот как это работает:

Сортировка: во-первых, строки сортируются по одному или нескольким столбцам. Эти столбцы указываются в ORDER BY в конструкции OVER.

- Присвоение рангов: каждой уникальной строке или группе строк, имеющих одинаковые значения в столбцах сортировки, присваивается ранг. Ранг начинается с 1.

- Одинаковые значения: если у нескольких строк одинаковые значения в столбцах сортировки, они получают одинаковый ранг. Например, если две строки занимают второе место, обе получают ранг 2.

- Пропуск рангов: после группы строк с одинаковым рангом, следующий ранг увеличивается на количество строк в этой группе. Например, если две строки имеют ранг 2, следующая строка получит ранг 4, а не 3.

- Продолжение сортировки: этот процесс продолжается до тех пор, пока не будут присвоены ранги всем строкам в наборе результатов.

- DENSE_RANK — возвращает ранг каждой строки. Но в отличие от функции RANK, она не пропускает ранги и после группы одинаковых значений ранг увеличивается на единицу, а не на количество строк. Например, если две строки имеют ранг 2, следующая строка получит ранг 3, а не 4.

```
MySQL 8.1
```

```
SELECT id,
	home_type,
	price,
	ROW_NUMBER() OVER(PARTITION BY home_type ORDER BY price) AS "row_number",
	RANK() OVER(PARTITION BY home_type ORDER BY price) AS "rank",
	DENSE_RANK() OVER(PARTITION BY home_type ORDER BY price) AS "dense_rank"
FROM Rooms;

```

```

```

idhome_typepricerow_numberrankdense_rank5Entire home/apt8011138Entire home/apt852224Entire home/apt8933319Entire home/apt9944448Entire home/apt11055549Entire home/apt11566625Entire home/apt12077715Entire home/apt12087742Entire home/apt12097711Entire home/apt1351010816Entire home/apt1401111928Entire home/apt15012121010Entire home/apt15013121045Entire home/apt15014121046Entire home/apt15015121030Entire home/apt18016161120Entire home/apt1901717126Entire home/apt20018181317Entire home/apt2151919142Entire home/apt22520201521Entire home/apt29921211637Private room3511129Private room4422234Private room5033331Private room5043332Private room5255433Private room5566526Private room607767Private room6087641Private room6899735Private room70101088Private room79111199Private room791211927Private room8013131023Private room8014131050Private room8015131012Private room8516161114Private room8517161113Private room8918181236Private room8919181224Private room11020201343Private room12021211422Private room13022221547Private room13023221544Private room13524241618Private room1402525171Private room1492626183Private room15027271939Private room15028271940Shared room40111

### Оконные функции смещения

Оконные функции смещения — это функции, которые позволяют перемещаться и обращаться к разным строкам в окне, относительно текущей строки, а также обращаться к значениям в начале или в конце окна.

- 
LAG — обращается к данным из предыдущих строк окна.
Имеет три аргумента: столбец, значение которого необходимо вернуть, количество строк для смещения (по умолчанию 1), значение, которое необходимо вернуть, если после смещения возвращается значение NULL.

- 
LEAD — обращается к данным из следующих строк. Аналогично LAG имеет 3 аргумента.

- 
FIRST_VALUE — возвращает первое значение в окне. В качестве аргумента принимает столбец, значение которого необходимо вернуть.

- 
LAST_VALUE — возвращает последнее значение в окне. В качестве аргумента принимает столбец, значение которого необходимо вернуть

При использовании ORDER BY рамки окна по умолчанию устанавливаются от начала партиции до текущей строки (RANGE BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW). Из-за этого LAST_VALUE будет возвращать значение текущей строки, а не последней строки всей партиции. Чтобы получить действительно последнее значение партиции, необходимо явно расширить границы окна: ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING.

```
MySQL 8.1
```

```
SELECT id,
	home_type,
	price,
	LAG(price) OVER(PARTITION BY home_type ORDER BY price) AS "lag",
	LAG(price, 2) OVER(PARTITION BY home_type ORDER BY price) AS "lag_2",
	LEAD(price) OVER(PARTITION BY home_type ORDER BY price) AS "lead",
	FIRST_VALUE(price) OVER(PARTITION BY home_type ORDER BY price) AS "first_value",
	LAST_VALUE(price) OVER(PARTITION BY home_type ORDER BY price ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING) AS "last_value"
FROM Rooms;

```

```

```

idhome_typepricelaglag_2leadfirst_valuelast_value5Entire home/apt80<NULL><NULL>858029938Entire home/apt8580<NULL>89802994Entire home/apt898580998029919Entire home/apt9989851108029948Entire home/apt11099891158029949Entire home/apt115110991208029925Entire home/apt1201151101208029915Entire home/apt1201201151208029942Entire home/apt1201201201358029911Entire home/apt1351201201408029916Entire home/apt1401351201508029928Entire home/apt1501401351508029910Entire home/apt1501501401508029945Entire home/apt1501501501508029946Entire home/apt1501501501808029930Entire home/apt1801501501908029920Entire home/apt190180150200802996Entire home/apt2001901802158029917Entire home/apt215200190225802992Entire home/apt2252152002998029921Entire home/apt299225215<NULL>8029937Private room35<NULL><NULL>443515029Private room4435<NULL>503515034Private room504435503515031Private room505044523515032Private room525050553515033Private room555250603515026Private room60555260351507Private room606055683515041Private room686060703515035Private room70686079351508Private room79706879351509Private room797970803515027Private room807979803515023Private room808079803515050Private room808080853515012Private room858080853515014Private room858580893515013Private room898585893515036Private room8989851103515024Private room11089891203515043Private room120110891303515022Private room1301201101303515047Private room1301301201353515044Private room1351301301403515018Private room140135130149351501Private room149140135150351503Private room1501491401503515039Private room150150149<NULL>3515040Shared room40<NULL><NULL><NULL>4040