# AniFlux API Documentation

This document serves as the master API reference for the **AniFlux** media resolver engine. 

---

## Base Path
All routes below are prefixed with `/api`.

---

## 1. Get Trending Anime
Returns a paginated list of currently trending anime.

* **Route**: `GET /api/trending`
* **Query Parameters**:
  * `page` (optional, integer, default: `1`): The page number to fetch.
  * `perPage` (optional, integer, default: `5`): Items per page.
* **Response**: `200 OK` (JSON array of `Media` objects)
  ```json
  [
    {
      "id": "153518",
      "type": "anime",
      "title": "Jujutsu Kaisen Season 2",
      "original_title": "Jujutsu Kaisen 2nd Season",
      "cover": "https://s4.anilist.co/file/anilistcdn/media/anime/cover/large/bx153518-JnF3Oa1wG1T5.jpg",
      "banner": "https://s4.anilist.co/file/anilistcdn/media/anime/banner/153518-7FjU2kF8aZ1b.jpg",
      "description": "The second season of Jujutsu Kaisen...",
      "score": 88,
      "genres": ["Action", "Fantasy", "Supernatural"],
      "status": "FINISHED",
      "season": "SUMMER",
      "season_year": 2023,
      "total_episodes": 23,
      "duration": 24,
      "next_airing": null
    }
  ]
  ```

---

## 2. Get Seasonal Anime
Lists anime titles matching a specific season and year.

* **Route**: `GET /api/seasonal`
* **Query Parameters**:
  * `page` (optional, integer, default: `1`): Page number.
  * `perPage` (optional, integer, default: `24`): Items per page.
  * `season` (optional, string): The season name (`WINTER`, `SPRING`, `SUMMER`, `FALL`).
  * `year` (optional, integer): The calendar year (e.g. `2026`).
* **Response**: `200 OK` (JSON array of `Media` objects)

---

## 3. Search Anime
Searches the database/resolver for matches matching a text query.

* **Route**: `GET /api/search`
* **Query Parameters**:
  * `q` (required, string): The search query text (e.g. `Frieren`).
  * `page` (optional, integer, default: `1`): Page number.
  * `perPage` (optional, integer, default: `24`): Items per page.
* **Response**: `200 OK` (JSON array of `Media` objects)

---

## 4. Get Anime by Genre
Lists titles matching one or more comma-separated genres.

* **Route**: `GET /api/genre`
* **Query Parameters**:
  * `genre` (required, string): Comma-separated list of genres (e.g. `Action,Sci-Fi`).
  * `page` (optional, integer, default: `1`): Page number.
  * `perPage` (optional, integer, default: `24`): Items per page.
* **Response**: `200 OK` (JSON array of `Media` objects)

---

## 5. Get Current Airing Anime
Retrieves anime that are currently airing in the current week.

* **Route**: `GET /api/airing`
* **Query Parameters**:
  * `page` (optional, integer, default: `1`): Page number.
  * `perPage` (optional, integer, default: `24`): Items per page.
* **Response**: `200 OK` (JSON array of `Media` objects)

---

## 6. Get Weekly Airing Schedule
Retrieves the weekly chronological list of airing episodes.

* **Route**: `GET /api/schedule`
* **Query Parameters**:
  * `page` (optional, integer, default: `1`): Page number.
  * `perPage` (optional, integer, default: `24`): Items per page.
* **Response**: `200 OK` (JSON array of `Episode` objects)
  ```json
  [
    {
      "id": 1042531,
      "number": "8",
      "title": "The Journey Begins",
      "air_date": 1716827400,
      "overview": "Frieren starts her journey to the North...",
      "image": "https://image.tmdb.org/t/p/w500/..."
    }
  ]
  ```

---

## 7. Get Individual Anime details
Retrieves full details for a single anime by its AniList ID.

* **Route**: `GET /api/{id}`
* **Path Parameters**:
  * `id` (required, integer): The AniList media ID.
* **Response**: `200 OK` (JSON `Media` object)

---

## 8. Get Episodes List
Retrieves regular and special/OVA episodes for a media title.

* **Route**: `GET /api/{id}/episodes`
* **Path Parameters**:
  * `id` (required, integer): The AniList media ID.
* **Query Parameters**:
  * `page` (optional, integer, default: `1`): Page number for regular episodes pagination.
  * `perPage` (optional, integer, default: `24`): Items per page.
* **Response**: `200 OK` (JSON `EpisodeList` wrapper)
  ```json
  {
    "episodes": [
      {
        "id": 204351,
        "number": "1",
        "title": "First Episode",
        "air_date": 1716820000,
        "overview": "Introduction to characters...",
        "image": "https://..."
      }
    ],
    "specials": [
      {
        "id": 204399,
        "number": "S1",
        "title": "Special Episode 1",
        "air_date": 1716920000,
        "overview": "Recap special...",
        "image": "https://..."
      }
    ],
    "total_count": 12
  }
  ```

---

## 9. Get Episode Sources (Magnet Links)
Searches Torznab/Jackett indexers for high-quality torrent stream sources for a specific episode, sorted by seeders descending.

* **Route**: `GET /api/{id}/episodes/{epNumber}/sources`
* **Path Parameters**:
  * `id` (required, integer): The AniList ID.
  * `epNumber` (required, string): Episode number (e.g. `"1"`, `"S1"`).
* **Response**: `200 OK` (JSON array of `Source` objects)
  ```json
  [
    {
      "title": "[SubsPlease] Frieren - 08 (1080p) [98BA8D72].mkv",
      "magnet_uri": "magnet:?xt=urn:btih:...",
      "seeders": 142,
      "leechers": 15,
      "size": 1450284910,
      "info_hash": "a4d3f7b..."
    }
  ]
  ```

---

## 10. Get Anime Recommendations
Generates 5 recommended anime titles based on a reference AniList ID.

* **Route**: `GET /api/{id}/recommendations`
* **Path Parameters**:
  * `id` (required, integer): The AniList media ID.
* **Query Parameters**:
  * `page` (optional, integer, default: `1`): Page number.
  * `perPage` (optional, integer, default: `5`): Items per page.
* **Response**: `200 OK` (JSON array of `Media` objects)
