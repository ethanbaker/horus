<!--
  Created by: Ethan Baker (contact@ethanbaker.dev)
  
  Adapted from:
    https://github.com/othneildrew/Best-README-Template/
Here are different preset "variables" that you can search and replace in this template.
-->

<div id="top"></div>


<!-- PROJECT SHIELDS/BUTTONS -->
![alpha](https://img.shields.io/badge/status-alpha-red)
[![License][license-shield]][license-url]
[![Issues][issues-shield]][issues-url]


[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![LinkedIn][linkedin-shield]][linkedin-url]

<!-- 
NEED GITHUB WORKFLOW [![Go Coverage](https://github.com/ethanbaker/horus/wiki/coverage.svg)](https://raw.githack.com/wiki/ethanbaker/horus/coverage.html)
-->

<h4>Utils</h4>

[![GoDoc](https://godoc.org/github.com/ethanbaker/horus/utils?status.svg)](https://godoc.org/github.com/ethanbaker/horus/utils)
[![Go Report Card](https://goreportcard.com/badge/github.com/ethanbaker/horus/utils)](https://goreportcard.com/report/github.com/ethanbaker/horus/utils)
[![Go Coverage Report](./docs/utils-coverage.svg)](#)

<h4>Bot</h4>

[![GoDoc](https://godoc.org/github.com/ethanbaker/horus/bot?status.svg)](https://godoc.org/github.com/ethanbaker/horus/bot)
[![Go Report Card](https://goreportcard.com/badge/github.com/ethanbaker/horus/bot)](https://goreportcard.com/report/github.com/ethanbaker/horus/bot)
[![Go Coverage Report](./docs/bot-coverage.svg)](#)

<h4>Outreach</h4>

[![GoDoc](https://godoc.org/github.com/ethanbaker/horus/outreach?status.svg)](https://godoc.org/github.com/ethanbaker/horus/outreach)
[![Go Report Card](https://goreportcard.com/badge/github.com/ethanbaker/horus/outreach)](https://goreportcard.com/report/github.com/ethanbaker/horus/outreach)
[![Go Coverage Report](./docs/outreach-coverage.svg)](#)

<h4>Implementation - Discord</h4>

[![GoDoc](https://godoc.org/github.com/ethanbaker/horus/implementations/discord?status.svg)](https://godoc.org/github.com/ethanbaker/horus/implementations/discord)
[![Go Report Card](https://goreportcard.com/badge/github.com/ethanbaker/horus/implementations/discord)](https://goreportcard.com/report/github.com/ethanbaker/horus/implementations/discord)

<!-- PROJECT LOGO -->
<br><br><br>
<div align="center">
  <a href="https://github.com/ethanbaker/horus">
    <img src="./docs/logo.png" alt="Logo" height="80">
  </a>

  <h3 align="center">Horus</h3>

  <p align="center">
    A personal assistant bot, created for all my needs
  </p>
</div>


<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li><a href="#getting-started">Getting Started</a></li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>


<!-- ABOUT -->
## About

Horus is my personal assistant bot for any and all purposes. The desire to create them came from a wish for a personal assistant that could tie into specific user information, such as emails, calendars, Notion databases, and more.

Horus currently has two main capacities: functioning as a GPT assistant (off of OpenAI's client and function calling, though with goals to eventually pivot to Llama3), and performing outreach to the user to give live updates/reminders.

These functions are achieved using implementations, which use basic Horus functionality in a specific user interface, such as a Discord bot, speech to text, or website. All communications occur through HTTP channels so the same base Horus runner can supply multiple implemenations with consistent information.

The `bot` directory contains a generic assistant template for Horus to run off of. Custom modules are utilized with OpenAI's tool functionality to handle custom functions. The main idea of this module is to take advantage of GPT assistants with easily maintainable, custom-made modules.

The `outreach` directory contains code used to reach out to the user unprompted, such as for reminders or alarms. Different types of 'messages' exist to contact the user and are defined as follows:
* **Dyanmic Messages**: used to check dynamic content repeatedly and send a message to the user (ex: check a schedule database every given interval and send the user a message when the start time lines up)
* **Static Messages**: used to message the user at a set, fixed time according to a cron string (ex: send a daily calendar digest every morning at 7:00)
* **Timed Messages**: used to message the user once after a given time (ex: send the user a reminder after 3 hours, or send the user a reminder on August 11th at 16:00)

The `implementations` directory contains different implementations of Horus. These implementations allow the user to interact with Horus, and can contain utility functions specific to the implementation type. Current implementations include
* Discord Bot
* Terminal Interface

<p align="right">(<a href="#top">back to top</a>)</p>


### Built With

* [OpenAI GPT Models](https://platform.openai.com/docs/models)
* [Golang](https://go.dev/)
* [GORM](https://github.com/go-gorm/gorm)
* [Discord Go](https://github.com/bwmarrin/discordgo)
* [Notion API](https://developers.notion.com/)

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- GETTING STARTED -->
## Getting Started

Horus is built by me, for me. A plethroa of generalized assistant bots exist to use for a wide variety of general purposes, but this bot was built with specific intent to provide me with specialized usecases. Because of this, Horus is not an "install right out of the box" kind of assistant. If you'd like to make use any of this functionality here, I'd recommend using the code as a template and custom-coding your own modules.

Here is a list of different sections of the code base and how they can be customized:
* `/bot`: add/modify custom modules according to the `/bot/template` directory. Implemented modules contain details about usage and credential needs
* `/outreach`: add/modify custom messages according to other examples
* `/implementations`: add/modify configuration setup to meet personal needs


<p align="right">(<a href="#top">back to top</a>)</p>


<!-- USAGE EXAMPLES -->
## Usage

Horus can be used as a personal assistant through many implementations. Pictures of example usage will be located here in the future.

_For more examples, please refer to the [documentation][documentation-url]._

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- ROADMAP -->
## Roadmap

- [x] Outreach
- [ ] Speech to text implementation
- [ ] Horus API
    - [ ] API Wrapper
    - [ ] Implementation API Usage

See the [open issues][issues-url] for a full list of proposed features (and known issues).

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- CONTRIBUTING -->
## Contributing

For issues and suggestions, please include as much useful information as possible.
Review the [documentation][documentation-url] and make sure the issue is actually
present or the suggestion is not included. Please share issues/suggestions on the
[issue tracker][issues-url].

For patches and feature additions, please submit them as [pull requests][pulls-url]. 
Please adhere to the [conventional commits][conventional-commits-url]. standard for
commit messaging. In addition, please try to name your git branch according to your
new patch. [These standards][conventional-branches-url] are a great guide you can follow.

You can follow these steps below to create a pull request:

1. Fork the Project
2. Create your Feature Branch (`git checkout -b branch_name`)
3. Commit your Changes (`git commit -m "commit_message"`)
4. Push to the Branch (`git push origin branch_name`)
5. Open a Pull Request

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- LICENSE -->
## License

This project uses the GNU General Public License.

You can find more information in the [LICENSE][license-url] file.

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- CONTACT -->
## Contact

Ethan Baker - contact@ethanbaker.dev - [LinkedIn][linkedin-url]

Project Link: [https://github.com/ethanbaker/horus][project-url]

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* [The Egyptian god Horus](https://en.wikipedia.org/wiki/Horus)

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/ethanbaker/horus.svg
[forks-shield]: https://img.shields.io/github/forks/ethanbaker/horus.svg
[stars-shield]: https://img.shields.io/github/stars/ethanbaker/horus.svg
[issues-shield]: https://img.shields.io/github/issues/ethanbaker/horus.svg
[license-shield]: https://img.shields.io/github/license/ethanbaker/horus.svg
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?logo=linkedin&colorB=555

[contributors-url]: <https://github.com/ethanbaker/horus/graphs/contributors>
[forks-url]: <https://github.com/ethanbaker/horus/network/members>
[stars-url]: <https://github.com/ethanbaker/horus/stargazers>
[issues-url]: <https://github.com/ethanbaker/horus/issues>
[pulls-url]: <https://github.com/ethanbaker/horus/pulls>
[license-url]: <https://github.com/ethanbaker/horus/blob/master/LICENSE>
[linkedin-url]: <https://linkedin.com/in/ethandbaker>
[project-url]: <https://github.com/ethanbaker/horus>

[product-screenshot]: path_to_demo
[documentation-url]: <https://godoc.org/github.com/ethanbaker/horus>

[conventional-commits-url]: <https://www.conventionalcommits.org/en/v1.0.0/#summary>
[conventional-branches-url]: <https://docs.microsoft.com/en-us/azure/devops/repos/git/git-branching-guidance?view=azure-devops>