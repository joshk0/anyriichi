In this project, we'll be creating a new web application called anyRiichi. The target audience is riichi mahjong players who use internet platforms such as Riichi City, Tenhou, and Mahjong Soul. The target audience is an English or Japanese speaker, for now. Its purpose combines what currently exists across a few different sites:

* Amae Koromo ([https\://amae-koromo.sapk.ch](https://amae-koromo.sapk.ch)) source available at [https\://github.com/SAPikachu/amae-koromo](https://github.com/SAPikachu/amae-koromo)  \- for Mahjong Soul  
* Nodocchi.moe ([https\://nodocchi.moe](https://nodocchi.moe)) \- no source available \- for Tenhou.net

In each case, the sites scrape data from the relevant Riichi Mahjong game platform, and present a player-centric view of performance. Depending on the platform, some data is easily available and some is not:

* Mahjong Soul: Amae Koromo seems to be able to use the 'Watch' functionality inside Mahjong Soul's websockets API to monitor all of the rooms that are monitorable, and wait for games to finish, then obtain their logs and calculate statistics.  
* Tenhou: [https\://tenhou.net/sc/raw/](https://tenhou.net/sc/raw/) is the source that Nodocchi uses; a job can be run periodically (every 20 minutes), according to Tenhou's acceptable usage guidelines, to obtain a listing of available game result files, then fetch the game result files that have had their file size changed since you last fetched them, or if they're files you don't have.  
* Note: No site like Amae Koromo exists for Riichi City yet \- so this is an opportunity for us to attempt to support it with our application.

The player-centric view of performance available on both sites (which you should research directly using pages like [https\://amae-koromo.sapk.ch/player/120105558/16](https://amae-koromo.sapk.ch/player/120105558/16) (for Mahjong Soul player LightPinkYoshi) and [https\://nodocchi.moe/tenhoulog/\#\!\&name=%E5%8F%A3%E3%83%B3.%E3%82%AD%E3%83%9B%E3%83%BC%E3%83%86](https://nodocchi.moe/tenhoulog/#!&name=%E5%8F%A3%E3%83%B3.%E3%82%AD%E3%83%9B%E3%83%BC%E3%83%86) (for Tenhou player 口ン.キホーテ).

* The recent matches played by the player are available, and can be filtered to the specific skill-based game lobby (Gold, Jade for Mahjong Soul, then Ippan, Joukyuu and Tokujou, for Mahjong Soul) and the lobby type (3 player or 4 player ranked, or special tournaments)  
* The match shows the results of the game \- the 3 or 4 players who were involved and their final point results from the game. In Mahjong Soul / Amae Koromo the score at the end of the game is shown, in Tenhou, the post-uma \+/- points are used.)

Our app should be able to do the following:

* A player should be able to register for the site, then add their user IDs for each platform we support. Target supported platforms should be Mahjong Soul and Tenhou first, next step Riichi City.  
* Using the sources available to us, we should be able to show them their personal page, with filter for platform, lobby, 4 player or 3 player, ranked or tournaments,  or any combination of the preceding.  
* The user should be able to see statistics like their placement percentages (% of being 1st place over all visible games, 2nd place, 3rd place, 4th place etc.)  
* Other than this, the stats available on the inspired sites, Amae Koromo and Nodocchi should be imported as well. The stats on Amae Koromo are richer because they are able to produce the stats from the full detailed game logs showing every single discard / call of every single hand of a given game, and create stats about what each player did in each game.  
* In the case of Tenhou, the game records don't include the full game log. The user should be able to submit them manually, and in the case of people using Tenhou in a browser, they should be able to use a browser extension to sniff the Local Application data for tenhou and automatically submit new logs to the user's account. In this way, we can provide a richer experience than Nodocchi as we can calculate Amae-Koromo type statistics but for Tenhou. There would be no solution for mobile users, but they should still be able to bulk submit logs (e.g. URLs like [https\://tenhou.net/0/?log=2026112505gm-0099-0000-39171f43\&tw=0](https://tenhou.net/0/?log=2026112505gm-0099-0000-39171f43&tw=0) )  
* Longer term, the user should be able to take any game from any platform, and feed it into known AI game analysis platforms, such as Bigcoach ( [https\://review.bigcoach.work/](https://review.bigcoach.work/) ) or NAGA ([https\://naga.dmv.nico/](https://naga.dmv.nico/) ). This would mean that you could take a Mahjong Soul log and feed it to NAGA, which only supports Tenhou logs; conversion would take place in the background, and if we couldn't submit it directly we could at least give the user a file to submit to the service.

Underlying the app, should be the following technologies:

* UI code should be written in TypeScript. A UI framework such as React or something more modern could be used on top of that, I'm not particular as to which.  
* Worker code could be written in Go, Python, or TypeScript as well. I'd prefer Go, unless there's a compelling reason to use something else.  
* A database, either SQL or NoSQL, should be used to store the game records across platforms, as well as the user list.  
* An authentication handler should be used, whatever can be done for free or a cheap cost, so that people can sign up, enable 2fa, use social logins, etc. I'd prefer not to roll my own or use an open source project for the same, unless there really is no free option without huge caveats.

We are agnostic on what platform you should build this on, but certainly we should use TypeScript for implementing any browser side code at a minimum, and then wherever is easiest to deploy, such as Vercel or Cloud Run. I'm open to other suggestions for places to deploy that make it easiest for agentic development as well. We can assume we could have a worker available (probably should be a freestanding codebase that could be deployed on its own) to connect to the same database and do the periodic jobs that we need to do to keep the data fresh.

We'll use Github projects to manage what will be a huge backlog for this project. You can use [https\://github.com/joshk0/anyriichi](https://github.com/joshk0/anyriichi) for this. There are a surely a lot of open questions. Ask as many questions as you need to and never design something without confirmation or proposing a mockup of what you want to build. Also, take this input spec and immortalize it into a directory of design documents in the anyriichi repository, so that you can refer to it later as you build everything out. It probably shouldn't be one single design document, but segmented into logical components of the application, such as:

* web application  
  * authentication  
  * user interface  
  * personas of users  
  * analysis of capabilities of related projects (amae koromo, nodocchi)  
* storage  
  * user model  
  * game record model  
  * statistics model  
* worker  
  * worker architecture  
* infrastructure  
  * infrastructure choices such as db, compute platform, scheduled task management, SDKs and libraries

Even if you don't get all possible questions answered right away, record the open questions into the design document hierarchy, and resolve the answer to the question when it becomes relevant for implementation.

Here are some other related projects that could be integrated with this one. Presence of a project here does not ratify that our app should do everything that each example does, but each example has traits or logic we may wish to extract for our own platform.

[https\://mjai.ekyu.moe/](https://mjai.ekyu.moe/)  \- another free analysis tool like bigcoach or NAGA  
[https\://github.com/honvl/Majsoul-to-NAGA](https://github.com/honvl/Majsoul-to-NAGA) \- convert majsoul log to format supported by NAGA  
[https\://zx-api.pesiu.org/](https://zx-api.pesiu.org/) \- riichi city log scraper? try log : dab6gskj4ak0jap8ge00@1  
[https\://github.com/Euophrys/tenhou-log](https://github.com/Euophrys/tenhou-log) \- fetching all tenhou logs for all time into a local analyzable format; knows how to extract the local application list of logs, for when we build a Chrome / Firefox extension to automatically submit them after the game finishes

Also, I can provide authentication information for every relevant platform in a .env file so you can connect as a real user and do things that a regular game player can, such as looking up arbitrary log files within a given mahjong platform  usually requires authentication, as well as log files / game IDs from each platform whenever necessary for testing and analysis.