package com.coooolfan.onlyboxes.app

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.json.JsonMapper
import com.coooolfan.onlyboxes.core.service.CodeExecutor
import com.coooolfan.onlyboxes.core.service.StatefulCodeExecutorService
import com.coooolfan.onlyboxes.infra.boxlite.BoxliteBoxFactory
import org.springframework.beans.factory.annotation.Value
import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication
import org.springframework.context.annotation.Bean

@SpringBootApplication(proxyBeanMethods = false)
class App {
    @Bean("mcpServerObjectMapper")
    fun mcpServerObjectMapper(): ObjectMapper {
        return JsonMapper
            .builder()
            .findAndAddModules()
            .build()
    }

    @Bean
    fun codeExecutor(
        @Value("\${onlyboxes.lease.min-seconds:30}")
        minLeaseSeconds: Long,
        @Value("\${onlyboxes.lease.max-seconds:3600}")
        maxLeaseSeconds: Long,
    ): CodeExecutor {
        return StatefulCodeExecutorService(
            boxFactory = BoxliteBoxFactory(),
            minLeaseSeconds = minLeaseSeconds,
            maxLeaseSeconds = maxLeaseSeconds,
        )
    }

    @Bean
    fun mcpController(
        codeExecutor: CodeExecutor,
        authTokenProvider: AuthTokenProvider,
    ): McpController {
        return McpController(
            codeExecutor = codeExecutor,
            authTokenProvider = authTokenProvider,
        )
    }

    companion object {
        @JvmStatic
        fun main(args: Array<String>) {
            runApplication<App>(*args)
        }
    }
}
