# Coding Tests 1

## Description

I applied for a startup job which had three small tests.  I completed them using Ruby, Rails, and Go (golang).

1. A script which scrapes the Alexa Top 100 site listings and stores them into a db ([website\_and\_scraper/lib/tasks/alexa.rake](https://github.com/dseevr/coding_tests_1/blob/master/website_and_scraper/lib/tasks/alexa.rake))
2. A web frontend which displays a paginated view of the listings and allows editing via a modal popup + AJAX (the Rails app under [website\_and\_scraper](https://github.com/dseevr/coding_tests_1/tree/master/website_and_scraper))
3. A URL shortener service: creates short URLs, redirects when one is visited, records some stats on visitors, and exposes stats for short URLs ([link_shortener/main.go](https://github.com/dseevr/coding_tests_1/blob/master/link_shortener/main.go))

## License

BSD
