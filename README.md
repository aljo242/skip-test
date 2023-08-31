# skip-test
skip take home test

## Running

To run, use the following command:

```shell
 make run
```

## Design

The overall flow of the code is to:
- Download the data:
  - Since each token is downloaded one at a time, we can also build a map of all usages of the attribute
keys used to determine rarity
  - This operation can be done in parallel since we can make as many parallel calls to the token endpoint
- Calculate the rarity scores:
  - Each token's rarity can be calculated independently once a map of attribute occurrences has been generated
  - This calculation is just an implementation of the given equation in the brief, using accesses from the sync map
- Sort the tokens once all rarities have been calculated:
  - For now the serial `slices.Sort` function is used to compare the token rarity values
  - This sorting could also be parallelized using one of many parallel sorting algorithms.  See [resource](https://www.massey.ac.nz/~mjjohnso/notes/59735/myslides8.pdf).

To ensure atomicity and a deterministic result, a small wrapper around the builtin `map` construct is used.
This wrapper struct has helper functions like `Load()` and `Store()` which add mutex lock/unlocks to ensure that there are no 
concurrent accesses by threads.

For parallelization, the number of threads is configurable using the `viper` package to read a `config.yaml` file.
If no config file is present, the default value of 1 thread will be used.

Note: 

To support odd numbers of threads or tokens, overflow is handled during parallelization.

Final Display using 10 threads:
```shell
2023/08/31 12:15:51  Displaying Top 5 Tokens... 
2023/08/31 12:15:51  ID: 6088, Rarity: 0.008560
 
2023/08/31 12:15:51  ID: 9605, Rarity: 0.008456
 
2023/08/31 12:15:51  ID: 4666, Rarity: 0.008315
 
2023/08/31 12:15:51  ID: 3491, Rarity: 0.007779
 
2023/08/31 12:15:51  ID: 9224, Rarity: 0.007547

```

