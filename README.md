# DKP Auction House

DKP Auction House is a PocketBase application for managing auctions and bids in a DKP (Dragon Kill Points) system.

## TODO List
- [ ] Add token healtcheck (compare tokens owned by users and tokens in transactions).

## Features

- Create and manage auctions
- Place bids on auctions
- Automatically finish auctions and determine winners
- Notify users about auction updates

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/dkp-auction.git
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
