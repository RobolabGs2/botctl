# botctl
CLI утилита для проведения соревнований между ботами в рамках курса CS253. Интеллектуальные системы

# Требования к боту
+ Первым аргументом командной строки принимает каким цветом он играет: `0` или `1`
+ `stdin` принимает ходы соперника
+ `stdout` пишет свой ход
+ `stderr` пишет всё остальное - логи, информацию для человека, состояние доски
+ `exit code`: победа-`0`, поражение-`1`, ничья-`2`

# Использование
```
Usage:
  botctl [flags] path/to/mybot1.exe path/to/mybot2.exe
  -r int
    	Количество раундов (default 1)
  -v int
    	0 - логи ботов не выводятся, 1 - выводятся логи первого, 2 - обоих ботов (default 1)
```
 
Порядок ботов влияет на вывод: в `stdout` будет писаться `stderr` первого бота, в `stderr` - второго.

По окончанию всех раундов будет выведен суммарный счёт. При смене раунда цвета меняются.