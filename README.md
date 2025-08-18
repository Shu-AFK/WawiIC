# WawiIC - JTL Wawi Item Combination Tool  

<p align="center">
  <img src="assets/banner.png" alt="WawiIC Banner" width="100%">
</p>

## Overview
WawiIC is a **Go-based GUI application** designed as an add-on for the **JTL Wawi** system. It simplifies the process of:
- **Querying and selecting items** from your inventory
- **Combining them into a parent article** (e.g., for size/color variations)
- **Automatically optimizing the parent item's description** using OpenAI's ChatGPT
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
1. **Clone the repository**:
   ```sh  
   git clone https://github.com/Shu-AFK/WawiIC.git  
   cd WawiIC  
   go mod init github.com/yourusername/WawiIC 
   go mod tidy
   ```

