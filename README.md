# WawiIC - JTL Wawi Item Combination Tool  

<p align="center">
  <img src="assets/banner.png" alt="WawiIC Banner" width="100%">
</p>

[🌐 Visit the website](https://shu-afk.github.io/WawiIC/) 

## Overview
WawiIC is a **Go-based GUI application** designed as an add-on for the **JTL Wawi** system. It simplifies the process of:
- **Querying and selecting items** from your inventory
- **Combining them into a parent article** (e.g., for size/color variations)
- **Automatically optimizing the parent item's description** using OpenAI's ChatGPT
- **Automatically creating an image containing the child items**
- **Creating the new parent item** and linking all selected items as children

Perfect for managing product variants efficiently!  

## Features
✔ **Intuitive GUI** – Easily browse and select items from your JTL Wawi database.  
✔ **Parent-Child Article Creation** – Combine items into a structured parent with variants.  
✔ **AI-Powered Descriptions** – Automatically generate optimized descriptions via OpenAI.  
✔ **Seamless JTL Integration** – Directly creates and updates items in your Wawi system via the API.  

## Prerequisites
- **JTL Wawi API** (tested on version [1.1])
- **Go** (≥ 1.24 recommended)
- **OpenAI API Key** (for description optimization. The token has to be saved in an environment variable named: "OPENAI_API_KEY") 
- **C Compiler (gcc)** (for Fyne GUI)

## Installation
1. Clone the repository and install dependencies: 
   ```sh  
   git clone https://github.com/Shu-AFK/WawiIC.git  
   cd WawiIC  
   go mod init github.com/Shu-AFK/WawiIC 
   go mod tidy
   
   go build -o WawiIC.exe
   ```

2. Configure your config according to [CONFIG.md](CONFIG.md) and place it in a folder called "config" in the same directory as the exe, or specify the path to the config file using the argument "-config {path to config.json}". 

3. Run the exe
