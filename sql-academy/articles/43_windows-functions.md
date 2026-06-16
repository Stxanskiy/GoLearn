# Оконные функции SQL

Оконные функции — мощный инструмент языка SQL, позволяющий проводить сложные вычисления по группам строк,
которые связаны с текущей строкой.

## Принцип работы

Возможно, вы зададитесь вопросом: «Что значит оконные?».
В стандартном SQL-запросе все наборы строк рассматриваются как один сплошной блок данных,
для которого и вычисляются агрегатные значения.
Однако, когда применяются оконные функции, запрос сегментируется на группы строк (или «окна»),
и для каждого такого сегмента подсчитываются индивидуальные агрегатные значения.
Это окно, которое подаётся в оконную функцию, может быть:

- всей таблицей

- отдельными партициями таблицы, то есть группой строк на основе одного или нескольких полей

- или даже конкретным диапазоном строк в пределах таблицы или партиции.
Например, мы можем определить окно, которое будет передаваться в оконную функцию,
как предыдущая + текущая строка таблицы. И тогда для каждой строки значение агрегатной функции будет
подсчитываться по-своему, так как данные, которые поступают в функцию, будут динамически меняться
от строки к строке. Окно будет как бы «скользить» по таблице.

### Визуализация

Оконные функции всегда принимают на вход окно данных, которое указывает пользователь, и возвращают результат в отдельный столбец.
Давайте рассмотрим как это может выглядеть. Для этого возьмём оконную функцию AVG для вычисления среднего значения и вот
такую небольшую таблицу:

А теперь давайте посмотрим, как оконная функция будет работать для разных переданных окон:

- 
Если в качестве окна указать всю таблицу, то для всех строк окно будет совпадать, и на вход функции AVG будет
поступать один и тот же набор данных, и, соответственно, результат будет одинаковый.

- 
Если в качестве окна указать партицию по полю home_type, то на вход функции AVG будет
поступать набор жилых помещений с одинаковым типом, и, соответственно, в результате в новой колонке будет
отображаться средняя стоимость по жилью, чей тип совпадает с типом у текущей строки таблицы.

- 
В качестве окна можно указать и более специфический набор строк. Например, окно можно определить как "предыдущая + текущая строка"
таблицы. Тогда это будет выглядеть следующим образом:

Стоит отметить, что для первой строки окно будет состоять только из 1-ой записи, так как предыдущей строки нет.

## Синтаксис оконной функции

```
MySQL 8.1
```

```
SELECT <оконная_функция>(<поле_таблицы>)
OVER (
      [PARTITION BY <столбцы_для_разделения>]
      [ORDER BY <столбцы_для_сортировки>]
      [ROWS|RANGE <определение_диапазона_строк>]
)

```

```

```

Где:

- <оконная_функция>(<поле_таблицы>) — используемая оконная функция. Например AVG(price).

- Далее следует OVER, который определяет окно (группу строк), которое будет передаваться в оконную функцию.
Если конструкцию OVER () оставить без параметров, то окном будет выступать вся таблица.

Далее внутри OVER следуют 3 необязательных параметра, с помощью которых можно гибко настраивать окно:

- с помощью PARTITION BY <столбцы_для_разделения> выборка делится на
непересекающиеся подмножества, где каждое подмножество содержит строки с одинаковыми значениями в одном или нескольких столбцах, образуются партиции.

- с помощью ORDER BY <столбцы_для_сортировки> устанавливается порядок строк внутри окна, особо важную роль играет в оконных функциях ранжирования.

- с помощью ROWS|RANGE <определение_диапазона_строк> формируются диапазоны строк. С помощью этого параметра можно указать, сколько строк брать до и после
текущей в окно.

На каждом из этих параметров мы подробнее остановимся в следующих статьях.

## Пример использования оконной функции

Давайте с помощью оконных функций попробуем получить список имён студентов и то, сколько человек у них в классе.

Для начала давайте получим список студентов и идентификатор класса, в котором они учатся:

```
MySQL 8.1
```

```
SELECT
    Student.first_name,
    Student.last_name,
    Student_in_class.class
FROM
    Student_in_class
JOIN
    Student ON Student_in_class.student = Student.id;

```

```

```

first_namelast_nameclassNikolajSokolov9VyacheslavEliseev9IvanEfremov9AnatolijZHdanov9GeorgijNoskov9ArtyomSergeev9ArinaEvseeva9AngelinaVoroncova9EkaterinaUstinova9RaisaLapina9LeonidIgnatov9SnezhanaSeliverstova9SemyonBiryukov9GeorgijBaranov8YUliyaVishnyakova8ValentinaBolshakova8LeonidKryukov8VladislavCvetkov8SnezhanaMorozova8LyubovBorisova8AnfisaKalashnikova8AnnaOsipova8KristinaMyasnikova8KristinaSmirnova8BorisSimonov7DmitrijTrofimov7YAkovRozhkov7FyodorDrozdov7GlebStrelkov7AngelinaLukina7NinaOdincova7ValeriyaNovikova7GrigorijKapustin7VitalijPanfilov7SvyatoslavTarasov6MatvejYAkushev6IlyaAlekseev6LyubovZaharova6PolinaSidorova6ElizavetaSamojlova6YUliyaAvdeeva6MatvejBogdanov6IlyaFilippov6DenisMel6SvyatoslavMuravyov6AnnaKulagina5ZHannaFokina5ValeriyaLapina5ValentinaSazonova5NataliyaMyasnikova5ViktoriyaMakarova5StanislavLazarev5GennadijOvchinnikov5RomanSHilov4TimurSubbotin4DanilaOsipov4ArinaSilina4NadezhdaZaharova4LarisaSHCHerbakova4AleksandraBelozyorova4NatalyaDavydova4MariyaFadeeva4YUrijMarkov3KirillSHubin3GrigorijKolobov3SemyonTrofimov3VasilijUstinov3ValentinaSHarova3LarisaSavina3GalinaOrekhova3ArinaSHarapova2ViktoriyaSergeeva2VasilijKrasilnikov2TimurRusakov2GlebNesterov2DenisMakarov2ElizavetaSHilova2VeraEvseeva1MargaritaKabanova1AngelinaLazareva1SemyonVoronov1InnokentijNekrasov1ArtyomNikitin1EgorBelyakov1
А теперь, чтобы вычислить, сколько учащихся учится в каждом из классов и вывести эту информацию в новую колонку,
мы можем применить оконную функцию:

```
MySQL 8.1
```

```
SELECT
    Student.first_name,
    Student.last_name,
    Student_in_class.class,
    COUNT(*) OVER (PARTITION BY Student_in_class.class) AS student_count_in_class
FROM
    Student_in_class
JOIN
    Student ON Student_in_class.student = Student.id;

```

```

```

first_namelast_nameclassstudent_count_in_classEgorBelyakov17ArtyomNikitin17InnokentijNekrasov17SemyonVoronov17AngelinaLazareva17MargaritaKabanova17VeraEvseeva17DenisMakarov27ArinaSHarapova27ViktoriyaSergeeva27VasilijKrasilnikov27TimurRusakov27GlebNesterov27ElizavetaSHilova27KirillSHubin38YUrijMarkov38GrigorijKolobov38SemyonTrofimov38ValentinaSHarova38LarisaSavina38GalinaOrekhova38VasilijUstinov38TimurSubbotin49RomanSHilov49DanilaOsipov49ArinaSilina49NadezhdaZaharova49LarisaSHCHerbakova49AleksandraBelozyorova49NatalyaDavydova49MariyaFadeeva49GennadijOvchinnikov58StanislavLazarev58ViktoriyaMakarova58NataliyaMyasnikova58ValentinaSazonova58ValeriyaLapina58ZHannaFokina58AnnaKulagina58IlyaFilippov611SvyatoslavMuravyov611DenisMel611MatvejBogdanov611YUliyaAvdeeva611ElizavetaSamojlova611PolinaSidorova611LyubovZaharova611IlyaAlekseev611MatvejYAkushev611SvyatoslavTarasov611NinaOdincova710BorisSimonov710DmitrijTrofimov710YAkovRozhkov710FyodorDrozdov710GlebStrelkov710AngelinaLukina710ValeriyaNovikova710GrigorijKapustin710VitalijPanfilov710AnnaOsipova811GeorgijBaranov811YUliyaVishnyakova811ValentinaBolshakova811LeonidKryukov811VladislavCvetkov811LyubovBorisova811AnfisaKalashnikova811SnezhanaMorozova811KristinaMyasnikova811KristinaSmirnova811VyacheslavEliseev913IvanEfremov913AnatolijZHdanov913GeorgijNoskov913ArtyomSergeev913ArinaEvseeva913AngelinaVoroncova913EkaterinaUstinova913RaisaLapina913LeonidIgnatov913SnezhanaSeliverstova913SemyonBiryukov913NikolajSokolov913

### Что делает наша оконная функция

Выражение PARTITION BY Student_in_class.class разделяет все строки таблицы на партиции по полю class.
Так, для каждой из строк в оконную функцию будут подаваться только те строки таблицы, где поле class
совпадает с полем class в текущей строке.
Функция COUNT же возвращает количество переданных в неё строк, тем самым мы и получаем сколько учащихся
учится в каждом из классов.

## Порядок выполнения оконных функций в SELECT

При использовании оконных функций важно понимать, в какой последовательности они будут исполняться. Так, как мы
можем увидеть на схеме ниже, окна отрабатывают предпоследним шагом, уже после фильтрации и группировки, но
перед финальной сортировкой результатов выборки.

## Заключение

В этой статье мы кратко рассмотрели понятие оконных функций, их возможности и практическую пользу.
В следующих статьях мы более подробно рассмотрим каждый аспект оконных функций.
И напоследок давайте проверим, все ли мы поняли:
Какое ключевое отличие между оконными функциями и агрегатными функциями с группировкой в SQL?
Оконные функции и агрегатные функции с группировкой выполняют одни и те же вычисления, но с использованием разного синтаксиса.Оконные функции вычисляются для каждой строки независимо, возвращая результат в отдельный столбец. Агрегатные функции с группировкой в свою очередь группируют строки и применяются к сформированным группам.В оконных функциях используется PARTITION BY, а в агрегатных функциях с группировкой — нет.