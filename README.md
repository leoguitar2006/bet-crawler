# bet-crawler - find your games to bet faster

### How it works ?

The crawler accesses the website https://www.academiadasapostasbrasil.com/ and tracks all games listed on the home page that have not yet started or are twenty minutes or less into the first half.

After that, the criteria is applied to only bring games where the sum of percentages of 0-0 results in the first half of the home team and the visiting team are less than **50%**.

The *LTD* column is also shown, displaying ***Sim*** if the sum of the **0-0** percentages for the entire game of the home team and the away team is less than **20%** and ***Não*** if it is greater than or equal.

### How do I run it ?

Just put the *bet-crawler.exe* file in some folder and run it from the *cmd.exe*. Here is an exemple, using a folder called *"bets"* on CMD (Windows):

```C:\bets\>bet-crawler```