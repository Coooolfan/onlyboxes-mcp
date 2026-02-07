package com.coooolfan.boxlites.app

import com.coooolfan.boxlites.core.service.CodeExecutor
import com.coooolfan.boxlites.core.service.StatefulCodeExecutorService
import com.coooolfan.boxlites.infra.boxlite.BoxliteBoxFactory

internal object RuntimeRegistry {
    val codeExecutor: CodeExecutor by lazy {
        StatefulCodeExecutorService(BoxliteBoxFactory())
    }
}
