package com.coooolfan.onlyboxes.app

import jakarta.servlet.FilterChain
import jakarta.servlet.ServletRequest
import jakarta.servlet.ServletResponse
import org.springframework.mock.web.MockHttpServletRequest
import org.springframework.mock.web.MockHttpServletResponse
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class TokenAuthFilterTest {
    @Test
    fun missingHeaderReturnsUnauthorized() {
        val filter = TokenAuthFilter(
            headerName = "X-Onlyboxes-Token",
            allowedTokens = setOf("abc123"),
            protectedEndpoint = "/mcp",
        )
        val request = MockHttpServletRequest("POST", "/mcp")
        val response = MockHttpServletResponse()
        val chain = TrackingFilterChain()

        filter.doFilter(request, response, chain)

        assertEquals(401, response.status)
        assertEquals("text/plain;charset=UTF-8", response.contentType)
        assertEquals("Unauthorized", response.contentAsString)
        assertFalse(chain.invoked)
    }

    @Test
    fun headerWithInvalidCharsReturnsUnauthorized() {
        val filter = TokenAuthFilter(
            headerName = "X-Onlyboxes-Token",
            allowedTokens = setOf("abc123"),
            protectedEndpoint = "/mcp",
        )

        listOf("Abc123", "abc_123", "abc 123").forEach { token ->
            val request = MockHttpServletRequest("POST", "/mcp").apply {
                addHeader("X-Onlyboxes-Token", token)
            }
            val response = MockHttpServletResponse()
            val chain = TrackingFilterChain()

            filter.doFilter(request, response, chain)

            assertEquals(401, response.status)
            assertFalse(chain.invoked)
        }
    }

    @Test
    fun tokenNotInAllowListReturnsUnauthorized() {
        val filter = TokenAuthFilter(
            headerName = "X-Onlyboxes-Token",
            allowedTokens = setOf("abc123"),
            protectedEndpoint = "/mcp",
        )
        val request = MockHttpServletRequest("POST", "/mcp").apply {
            addHeader("X-Onlyboxes-Token", "dev01")
        }
        val response = MockHttpServletResponse()
        val chain = TrackingFilterChain()

        filter.doFilter(request, response, chain)

        assertEquals(401, response.status)
        assertFalse(chain.invoked)
    }

    @Test
    fun validTokenInAllowListPassesThrough() {
        val filter = TokenAuthFilter(
            headerName = "X-Onlyboxes-Token",
            allowedTokens = setOf("abc123", "dev01"),
            protectedEndpoint = "/mcp",
        )
        val request = MockHttpServletRequest("POST", "/mcp").apply {
            addHeader("X-Onlyboxes-Token", "dev01")
        }
        val response = MockHttpServletResponse()
        val chain = TrackingFilterChain()

        filter.doFilter(request, response, chain)

        assertEquals(200, response.status)
        assertTrue(chain.invoked)
    }

    @Test
    fun nonMcpPathBypassesAuth() {
        val filter = TokenAuthFilter(
            headerName = "X-Onlyboxes-Token",
            allowedTokens = setOf("abc123"),
            protectedEndpoint = "/mcp",
        )
        val request = MockHttpServletRequest("GET", "/health")
        val response = MockHttpServletResponse()
        val chain = TrackingFilterChain()

        filter.doFilter(request, response, chain)

        assertEquals(200, response.status)
        assertTrue(chain.invoked)
    }

    @Test
    fun emptyAllowListRejectsRequests() {
        val filter = TokenAuthFilter(
            headerName = "X-Onlyboxes-Token",
            allowedTokens = emptySet(),
            protectedEndpoint = "/mcp",
        )
        val request = MockHttpServletRequest("POST", "/mcp").apply {
            addHeader("X-Onlyboxes-Token", "abc123")
        }
        val response = MockHttpServletResponse()
        val chain = TrackingFilterChain()

        filter.doFilter(request, response, chain)

        assertEquals(401, response.status)
        assertFalse(chain.invoked)
    }

    @Test
    fun invalidTokenInConfigThrows() {
        val ex = assertFailsWith<IllegalStateException> {
            parseAllowedTokens("abc123,abc-1")
        }

        assertTrue(ex.message.orEmpty().contains("^[a-z0-9]+$"))
    }

    @Test
    fun customHeaderNameWorks() {
        val filter = TokenAuthFilter(
            headerName = "X-Api-Token",
            allowedTokens = setOf("abc123"),
            protectedEndpoint = "/mcp",
        )
        val request = MockHttpServletRequest("POST", "/mcp").apply {
            addHeader("X-Api-Token", "abc123")
        }
        val response = MockHttpServletResponse()
        val chain = TrackingFilterChain()

        filter.doFilter(request, response, chain)

        assertEquals(200, response.status)
        assertTrue(chain.invoked)
    }

    private class TrackingFilterChain : FilterChain {
        var invoked: Boolean = false

        override fun doFilter(request: ServletRequest, response: ServletResponse) {
            invoked = true
        }
    }
}
