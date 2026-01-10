# portfolio-performance

Original prject by enrichman ((https://github.com/enrichman/portfolio-performance/)

This repository generates static JSON files with securities quotes from different sources. These quotes can be added easily in Portfolio Performance.

They are updated daily and available under the `https://shrike78.github.io/portfolio-performance/json/<ISIN>.json` URL.

Example: https://shrike78.github.io/portfolio-performance/json/XS1967674521.json

## Add a quote

If a quote is not present it needs to be added in the [`securities.csv`](https://github.com/shrike78/portfolio-performance/blob/main/securities.csv). It needs the ISIN, a Name, and a "loader". If the loader does not exists already it needs to be implemented.

## How To add a quote to Portfolio Performance

Add an empty instrument and add the JSON historical quotes.

To load the quotes in Portfolio Performance just add the URL of the quotes that you need with the proper JSONPath expression:

- Date: `$[*].date`
- Close: `$[*].close`

<img width="752" alt="screenshot_1" src="https://github.com/user-attachments/assets/dcca6929-f588-4e8e-8e0e-d53702644133" />

You will be able to see the graph in the main screen now:

<img width="1429" alt="Screen Shot 2023-03-30 at 14 40 48" src="https://user-images.githubusercontent.com/1763949/228841568-af1baae0-7228-4bd5-bf2d-7a01a9c092ae.png">
