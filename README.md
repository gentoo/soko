<p align="center"><img width=15% src="https://upload.wikimedia.org/wikipedia/commons/thumb/4/4f/Larry-the-cow-full.svg/1200px-Larry-the-cow-full.svg.png"></p>
<!--
<p align="center"><img width=60% src="https://packages.gentoo.org/assets/pgo-label.png"></p>
-->
<p align="center">
<a href="https://gitlab.com/gentoo/soko/-/pipelines"> <img src="https://gitlab.com/gentoo/soko/badges/master/pipeline.svg"></a>
<a href="https://blog.golang.org/go1.14" ><img src="https://img.shields.io/badge/Go-v1.14-blue"></a>
<a href="#contributing"> <img src="https://img.shields.io/badge/contributions-welcome-orange.svg"></a>
<a href="https://packagestest.gentoo.org"><img src="https://img.shields.io/badge/staging%20environment-develop-blue" /></a>
<a href="https://opensource.org/licenses/TBD"><img src="https://img.shields.io/badge/license-TBD-blue.svg"></a>
</p>


## tl;dr

This the code that powers [packages.gentoo.org](https://packages.gentoo.org/), internally codenamed soko which is Korean for package (who would have thought!)

## Table Of Contents 

- [Usage](#usage)
- [Contributing](#contributing)
- [History](#history)
- [License](#license)

## Usage

To get started quickly you can use docker-compose: 
```
$ docker-compose up
```


## Contributing

There are different ways to contribute:
  * Email [gpackages@gentoo.org](mailto:gpackages@gentoo.org) or
  * file a bug on [bugs.gentoo.org](https://bugs.gentoo.org/) (Websites → Packages) or
  * file a pull request on [github](https://github.com/gentoo/soko)

## History

Kkuleomi is at least the sixth rewrite of packages.gentoo.org.
Some of the rewrites were complete flops, and never went public.



### 2020-present: 'soko'
* https://gitweb.gentoo.org/sites/soko.git/
* Golang
* PostgreSQL backend.
* Authors:
   * [Max Magorsch](mailto:arzano@gentoo.org)
* Contributors:
   * [Alec Warner (antarus)](mailto:antarus@gentoo.org)

### 2016-2019: 'kkuleomi'
* https://gitweb.gentoo.org/sites/packages.git/
* Ruby on Rails
* ElasticSearch backend.
* Authors:
   * [Alex Legler (a3li)](mailto:a3li@gentoo.org)
* Contributors:
   * [Alex Legler (a3li)](mailto:a3li@gentoo.org) (2016)
   * [Robin H. Johnson (robbat2)](mailto:robbat2@gentoo.org) (2017-2020)
   * [Alec Warner (antarus)](mailto:antarus@gentoo.org) (2018-2020)
   * [Hans de Graaff (graaff)](mailto:graaff@gentoo.org) (2019-2020)
   * [Max Magorsch](mailto:arzano@gentoo.org) (2019-2020)

### 2012: 'gentoo-packages' (never deployed)
* https://gitweb.gentoo.org/proj/gentoo-packages.git/
* Never launched
* GSOC2012 rewrite
* Python & Django
* Authors:
   * [Slava Bacherikov](mailto:)

### 2007-2015:
* https://gitweb.gentoo.org/packages.git/
* Runs in production, 2007-2015.
* Python, CherryPy & Genshi
* MySQL backend
* Authors:
   * [Markus Ullmann (jokey)](mailto:jokey@gentoo.org) (2007)
   * [Robin H. Johnson (robbat2)](mailto:robbat2@gentoo.org) (2007-)
* Contributors:
   * [Alec Warner (antarus)](mailto:antarus@gentoo.org)
   * [Christian Ruppert (idl0r)](mailto:idl0r@gentoo.org)
   * [John Klehm](mailto:xixsimplicityxix@gmail.com)
   * [Pavlos Ratis (dastergon)](mailto:dastergon@gentoo.org)

### 2005-2007: 'P2' (never deployed)
* https://sources.gentoo.org/cgi-bin/viewvc.cgi/gentoo/src/packages/?pathrev=pre_2-0
* CVS branch `pre_2-0`
* Never launched.
* Python, Quixote (http://www.mems-exchange.org/software/quixote/)
* MySQL backend
* Authors:
   * [Albert Hopkins (marduk)](mailto:marduk@gentoo.org)
* Contributors: (unknown)

### 2004-2007
* https://sources.gentoo.org/cgi-bin/viewvc.cgi/gentoo/src/packages/
* first known `packages.gentoo.org` codebase
* Runs in production 2004 - mid-2007.
* Generate static HTML with use of server-side includes.
* Python, no framework.
* MySQL backend
* Authors:
   * [Albert Hopkins (marduk)](mailto:marduk@gentoo.org)
* Contributors: (unknown)

## License