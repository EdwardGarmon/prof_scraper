# prof_scraper
web scraper written in go to 
get data from rate my professor to use for 
a chrome extension


Uses two different strategies to get professor data.

The first strat is to directly query rate my professor. 
Rate my professor is so bad at indexing their website that this 
doesn't always work

If the first method doesn't work we then try a google search.
This works 99% of the time however google gets mad if your query them 
too much so trying rate my professor first avoids getting captcha'ed.
