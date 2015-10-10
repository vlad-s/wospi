"""
wospi 0.1

word spider whose sole purpose is to crawl for strings
and generate a wordlist
"""

__author__ = "Vlad <vlad at vlads dot me>"
__version__ = "0.1"
__license__ = "GPL v2"

# pylint: disable=import-error
# pylint can't find BeautifulSoup (installed with pip)

import requests
import argparse
from threading import Thread
from bs4 import BeautifulSoup

class WordSpider(object):
    """ Main class """

    def __init__(self, output, url):
        self.min_length = 4
        self.user_agent = "wospi (v0.1) word spiderbro"
        self.with_strip = False
        self.output = output
        self.url = url

        self.data_dict = {"words": [], "urls": [], "strip": ".,\"'"}

        try:
            self.outfile = open(self.output, "w")
        except IOError:
            print "Can't write the file. Do you have write access?"
            exit(1)

    def url_magic(self, url, depth):
        """ Do the URL boogie all night long """

        domain = self.url.split("/")[0]+"//"+self.url.split("/")[2]

        if url.startswith("/"):
            crawl_url = domain+url
        elif url.startswith(domain):
            crawl_url = url
        else:
            return

        if crawl_url not in self.data_dict.get("urls"):
            self.data_dict.get("urls").append(crawl_url)
            link_worker = Thread(target=self.request,
                                 args=(crawl_url, int(depth)-1))
            link_worker.start()

    def request(self, url, depth):
        """ Do request, get content, spread the word """

        if depth < 0:
            exit(1)

        if url.startswith("/"):
            url_split = url.split("/")
            url = url_split[0] + "//" + url_split[2]

        print "[+] URL: %s" % url

        headers = {"user-agent": self.user_agent}
        try:
            req = requests.get(url, headers=headers, timeout=3)
        except requests.ConnectionError:
            print "[+] Connection error, returning."
            return
        except requests.HTTPError:
            print "[+] Invalid HTTP response, returning."
            return
        except requests.Timeout:
            print "[+] Request timed out, returning."
            return
        except requests.TooManyRedirects:
            print "[+] Too many redirections, returning."
            return

        if "text/html" not in req.headers.get("content-type"):
            print "[+] Content type is not text/html, returning."
            return

        soup = BeautifulSoup(req.text, "html.parser")
        for invalid_tags in soup(["script", "iframe", "style"]):
            invalid_tags.extract()

        for link in soup.find_all("a"):
            if not isinstance(link.get("href"), type(None)):
                self.url_magic(link.get("href"), depth)

        data_worker = Thread(target=self.parse_data,
                             args=(soup.get_text(), ))
        data_worker.start()

    def parse_data(self, data):
        """ Parse the data after request """
        data = data.replace("\r\n", " ").replace("\n", " ").split()

        for word in data:
            word = word.encode("utf-8")
            if word not in self.data_dict.get("words"):
                if len(word) >= self.min_length:
                    if self.with_strip == True:
                        stripped = word
                        for char in self.data_dict.get("strip"):
                            stripped = stripped.strip(char)
                    self.data_dict.get("words").append(word)
                    self.outfile.write(word+"\n")
                    if self.with_strip == True and stripped != word:
                        self.data_dict.get("words").append(stripped)
                        self.outfile.write(stripped+"\n")

    def run(self, depth=0):
        """ Run, scraper, run! """
        self.request(self.url, depth)

if __name__ == "__main__":
    PARSER = argparse.ArgumentParser(description="word scraper/wordlist\
                                     generator")
    PARSER.add_argument("--min-length", type=int, default=4, help="minimum\
                        word length, defaults to 4")
    PARSER.add_argument("--user-agent", help="user agent to use on requests")
    PARSER.add_argument("--with-strip", action="store_true", help="also store\
                        the stripped word")
    PARSER.add_argument("--write", "-w", required=True, dest="file",
                        help="file to write the content in")
    PARSER.add_argument("--depth", default=0, help="crawling depth, defaults\
                        to 0")
    PARSER.add_argument("url", type=str, help="url to scrape")

    ARGS = PARSER.parse_args()

    SCRAPER = WordSpider(ARGS.file, ARGS.url)

    if ARGS.min_length is not None:
        SCRAPER.min_length = ARGS.min_length
    if ARGS.user_agent is not None:
        SCRAPER.user_agent = ARGS.user_agent
    if ARGS.with_strip == True:
        SCRAPER.with_strip = True

    SCRAPER.run(ARGS.depth)
