#!/usr/bin/env python3
"""
AIGateway MCP Server - Model management via FastMCP
Lists available models and manages custom model mappings
"""

import os
import httpx
import logging
from typing import Optional
from fastmcp import FastMCP

# Configure logging to stderr (required for MCP stdio transport)
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# API configuration
AIGATEWAY_URL = os.getenv("AIGATEWAY_URL", "http://localhost:8088")
AIGATEWAY_API_KEY = os.getenv("AIGATEWAY_API_KEY", "")

# Initialize FastMCP server
mcp = FastMCP("aigateway-mcp")


class AIGatewayClient:
    """Client for AIGateway HTTP API"""

    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url
        self.api_key = api_key
        self.client = httpx.Client(
            base_url=base_url,
            headers={"X-API-Key": api_key} if api_key else {}
        )

    def list_models(self) -> dict:
        """List all available models and custom mappings"""
        try:
            response = self.client.get("/v1/models")
            response.raise_for_status()
            return response.json()
        except httpx.HTTPError as e:
            logger.error(f"Failed to list models: {e}")
            raise

    def get_custom_mappings(self, user_id: str, api_key: str) -> list:
        """Get user's custom model mappings"""
        try:
            headers = {"X-API-Key": api_key}
            response = httpx.get(
                f"{self.base_url}/api/v1/model-mappings",
                headers=headers
            )
            response.raise_for_status()
            return response.json().get("data", [])
        except httpx.HTTPError as e:
            logger.error(f"Failed to get custom mappings: {e}")
            raise

    def create_mapping(self, api_key: str, alias: str, provider_id: str,
                      model_name: str, description: str = "") -> dict:
        """Create a new custom model mapping"""
        try:
            headers = {"X-API-Key": api_key}
            payload = {
                "alias": alias,
                "provider_id": provider_id,
                "model_name": model_name,
            }
            if description:
                payload["description"] = description

            response = httpx.post(
                f"{self.base_url}/api/v1/model-mappings",
                json=payload,
                headers=headers
            )
            response.raise_for_status()
            return response.json()
        except httpx.HTTPError as e:
            logger.error(f"Failed to create mapping: {e}")
            raise

    def update_mapping(self, api_key: str, alias: str,
                      provider_id: Optional[str] = None,
                      model_name: Optional[str] = None,
                      description: Optional[str] = None,
                      enabled: Optional[bool] = None) -> dict:
        """Update an existing custom model mapping"""
        try:
            headers = {"X-API-Key": api_key}
            payload = {}
            if provider_id:
                payload["provider_id"] = provider_id
            if model_name:
                payload["model_name"] = model_name
            if description is not None:
                payload["description"] = description
            if enabled is not None:
                payload["enabled"] = enabled

            response = httpx.put(
                f"{self.base_url}/api/v1/model-mappings/{alias}",
                json=payload,
                headers=headers
            )
            response.raise_for_status()
            return response.json()
        except httpx.HTTPError as e:
            logger.error(f"Failed to update mapping: {e}")
            raise


# Initialize client
client = AIGatewayClient(AIGATEWAY_URL, AIGATEWAY_API_KEY)


# Tool 1: List available models
@mcp.tool()
async def list_models(api_key: str) -> str:
    """
    List all available AI models including custom user-owned mappings.

    Args:
        api_key: User's API key from AIGateway (ak_...)

    Returns:
        JSON string with list of available models
    """
    try:
        logger.info(f"Listing models for user with API key: {api_key[:12]}...")

        # Get built-in models
        models_response = client.list_models()
        built_in_models = models_response.get("models", [])

        # Get custom mappings
        custom_mappings = client.get_custom_mappings("", api_key)

        # Format response
        result = {
            "built_in_models": built_in_models,
            "custom_mappings": custom_mappings,
            "total": len(built_in_models) + len(custom_mappings)
        }

        logger.info(f"Returned {result['total']} models")
        return str(result)

    except Exception as e:
        error_msg = f"Failed to list models: {str(e)}"
        logger.error(error_msg)
        return f"Error: {error_msg}"


# Tool 2: Create custom model mapping
@mcp.tool()
async def create_mapping(
    api_key: str,
    alias: str,
    provider_id: str,
    model_name: str,
    description: str = ""
) -> str:
    """
    Create a new custom model mapping (user-owned).

    Args:
        api_key: User's API key from AIGateway (ak_...)
        alias: Unique model alias (e.g., 'my-gpt')
        provider_id: Provider (antigravity, openai, or glm)
        model_name: Actual model name at the provider
        description: Optional description of the mapping

    Returns:
        JSON string with created mapping details
    """
    try:
        logger.info(f"Creating mapping: alias={alias}, provider={provider_id}")

        # Validate provider
        valid_providers = {"antigravity", "openai", "glm"}
        if provider_id not in valid_providers:
            return f"Error: Invalid provider_id. Must be one of: {valid_providers}"

        result = client.create_mapping(
            api_key=api_key,
            alias=alias,
            provider_id=provider_id,
            model_name=model_name,
            description=description
        )

        logger.info(f"Mapping created successfully: {alias}")
        return str(result)

    except Exception as e:
        error_msg = f"Failed to create mapping: {str(e)}"
        logger.error(error_msg)
        return f"Error: {error_msg}"


# Tool 3: Update custom model mapping
@mcp.tool()
async def update_mapping(
    api_key: str,
    alias: str,
    new_alias: str = "",
    provider_id: str = "",
    model_name: str = "",
    description: str = "",
    enabled: Optional[bool] = None
) -> str:
    """
    Update an existing custom model mapping (user-owned only).

    Args:
        api_key: User's API key from AIGateway (ak_...)
        alias: Current model alias to update
        new_alias: New alias (optional)
        provider_id: New provider (optional)
        model_name: New model name (optional)
        description: New description (optional)
        enabled: Enable/disable mapping (optional)

    Returns:
        JSON string with updated mapping details
    """
    try:
        logger.info(f"Updating mapping: alias={alias}")

        result = client.update_mapping(
            api_key=api_key,
            alias=alias,
            provider_id=provider_id if provider_id else None,
            model_name=model_name if model_name else None,
            description=description if description else None,
            enabled=enabled
        )

        logger.info(f"Mapping updated successfully: {alias}")
        return str(result)

    except Exception as e:
        error_msg = f"Failed to update mapping: {str(e)}"
        logger.error(error_msg)
        return f"Error: {error_msg}"


def main():
    """Run the MCP server"""
    logger.info(f"Starting AIGateway MCP Server")
    logger.info(f"AIGateway URL: {AIGATEWAY_URL}")
    logger.info(f"API Key configured: {bool(AIGATEWAY_API_KEY)}")

    # Run MCP server with stdio transport
    mcp.run(transport="stdio")


if __name__ == "__main__":
    main()
