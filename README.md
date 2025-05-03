[![Go Report Card](https://goreportcard.com/badge/github.com/oldbear24/dkp-auction)](https://goreportcard.com/report/github.com/oldbear24/dkp-auction)
# DKP Auction House

DKP Auction House is a PocketBase application for managing auctions and bids in a DKP (Dragon Kill Points) system.

## Features

- Create and manage auctions
- Place bids on auctions
- Automatically finish auctions and determine winners
- Notify users about auction updates

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/oldbear24/dkp-auction.git
    cd dkp-auction
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

3. Run the application:
    ```sh
    go run . serve
    ```
