package com.smkniba.kiosk

import android.Manifest
import android.annotation.SuppressLint
import android.content.Context
import android.content.pm.PackageManager
import android.graphics.Bitmap
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.view.View
import android.view.WindowManager
import android.webkit.PermissionRequest
import android.webkit.ValueCallback
import android.webkit.WebChromeClient
import android.webkit.WebResourceError
import android.webkit.WebResourceRequest
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import android.widget.EditText
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import com.smkniba.kiosk.databinding.ActivityMainBinding

class MainActivity : AppCompatActivity() {

    private lateinit var binding: ActivityMainBinding
    private var pendingPermissionRequest: PermissionRequest? = null

    private val prefName = "smk_niba_kiosk_prefs"
    private val prefKeyUrl = "server_url"

    // Runtime Camera Permission Launcher
    private val cameraPermissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { isGranted ->
        if (isGranted) {
            pendingPermissionRequest?.let { request ->
                request.grant(request.resources)
                pendingPermissionRequest = null
            } ?: run {
                loadServerUrl()
            }
        } else {
            Toast.makeText(
                this,
                getString(R.string.camera_permission_required),
                Toast.LENGTH_LONG
            ).show()
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Keep Screen On for Continuous Terminal Kiosk Usage
        window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)

        binding = ActivityMainBinding.inflate(layoutInflater)
        setContentView(binding.root)

        enableImmersiveMode()
        setupWebView()
        setupUIControls()

        checkAndRequestCameraPermission()
    }

    override fun onResume() {
        super.onResume()
        enableImmersiveMode()
    }

    private fun enableImmersiveMode() {
        WindowCompat.setDecorFitsSystemWindows(window, false)
        val controller = WindowInsetsControllerCompat(window, binding.root)
        controller.hide(WindowInsetsCompat.Type.systemBars())
        controller.systemBarsBehavior =
            WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
    }

    @SuppressLint("SetJavaScriptEnabled")
    private fun setupWebView() {
        binding.webView.apply {
            settings.apply {
                javaScriptEnabled = true
                domStorageEnabled = true
                databaseEnabled = true
                allowFileAccess = true
                allowContentAccess = true
                mediaPlaybackRequiresUserGesture = false
                useWideViewPort = true
                loadWithOverviewMode = true
                mixedContentMode = WebSettings.MIXED_CONTENT_ALWAYS_ALLOW
            }

            webChromeClient = object : WebChromeClient() {
                // Grant Camera Permission to HTML5 getUserMedia()
                override fun onPermissionRequest(request: PermissionRequest?) {
                    request?.let { req ->
                        val cameraResources = req.resources.filter {
                            it == PermissionRequest.RESOURCE_VIDEO_CAPTURE
                        }

                        if (cameraResources.isNotEmpty()) {
                            if (ContextCompat.checkSelfPermission(
                                    this@MainActivity,
                                    Manifest.permission.CAMERA
                                ) == PackageManager.PERMISSION_GRANTED
                            ) {
                                req.grant(req.resources)
                            } else {
                                pendingPermissionRequest = req
                                cameraPermissionLauncher.launch(Manifest.permission.CAMERA)
                            }
                        } else {
                            req.grant(req.resources)
                        }
                    }
                }

                override fun onProgressChanged(view: WebView?, newProgress: Int) {
                    if (newProgress < 100) {
                        binding.progressBar.visibility = View.VISIBLE
                        binding.progressBar.progress = newProgress
                    } else {
                        binding.progressBar.visibility = View.GONE
                    }
                }
            }

            webViewClient = object : WebViewClient() {
                override fun onPageStarted(view: WebView?, url: String?, favicon: Bitmap?) {
                    super.onPageStarted(view, url, favicon)
                    binding.layoutError.visibility = View.GONE
                }

                override fun onPageFinished(view: WebView?, url: String?) {
                    super.onPageFinished(view, url)
                    binding.swipeRefresh.isRefreshing = false
                }

                override fun onReceivedError(
                    view: WebView?,
                    request: WebResourceRequest?,
                    error: WebResourceError?
                ) {
                    super.onReceivedError(view, request, error)
                    if (request?.isForMainFrame == true) {
                        binding.swipeRefresh.isRefreshing = false
                        binding.layoutError.visibility = View.VISIBLE
                    }
                }
            }
        }
    }

    private fun setupUIControls() {
        // Swipe to Refresh
        binding.swipeRefresh.setOnRefreshListener {
            binding.webView.reload()
        }

        // Retry Button
        binding.btnRetry.setOnClickListener {
            binding.layoutError.visibility = View.GONE
            loadServerUrl()
        }

        // Settings Dialog Button
        binding.btnSettings.setOnClickListener {
            showServerUrlDialog()
        }

        // Secret long press on error title to open URL settings
        binding.tvErrorDetail.setOnLongClickListener {
            showServerUrlDialog()
            true
        }
    }

    private fun checkAndRequestCameraPermission() {
        if (ContextCompat.checkSelfPermission(
                this,
                Manifest.permission.CAMERA
            ) != PackageManager.PERMISSION_GRANTED
        ) {
            cameraPermissionLauncher.launch(Manifest.permission.CAMERA)
        } else {
            loadServerUrl()
        }
    }

    private fun getSavedServerUrl(): String {
        val prefs = getSharedPreferences(prefName, Context.MODE_PRIVATE)
        return prefs.getString(prefKeyUrl, getString(R.string.default_server_url))
            ?: getString(R.string.default_server_url)
    }

    private fun saveServerUrl(url: String) {
        val prefs = getSharedPreferences(prefName, Context.MODE_PRIVATE)
        prefs.edit().putString(prefKeyUrl, url.trim()).apply()
    }

    private fun loadServerUrl() {
        val url = getSavedServerUrl()
        binding.webView.loadUrl(url)
    }

    private fun showServerUrlDialog() {
        val currentUrl = getSavedServerUrl()
        val input = EditText(this).apply {
            setText(currentUrl)
            setSelection(currentUrl.length)
            hint = getString(R.string.hint_server_url)
            setPadding(48, 32, 48, 32)
        }

        AlertDialog.Builder(this)
            .setTitle(getString(R.string.title_settings_dialog))
            .setMessage("Masukkan URL Server Presensi Wajah:")
            .setView(input)
            .setPositiveButton(getString(R.string.btn_save)) { _, _ ->
                val newUrl = input.text.toString().trim()
                if (newUrl.isNotEmpty()) {
                    saveServerUrl(newUrl)
                    loadServerUrl()
                    Toast.makeText(this, "URL Berhasil Diperbarui!", Toast.LENGTH_SHORT).show()
                }
            }
            .setNegativeButton(getString(R.string.btn_cancel), null)
            .show()
    }

    @Deprecated("Deprecated in Java")
    override fun onBackPressed() {
        if (binding.webView.canGoBack()) {
            binding.webView.goBack()
        } else {
            super.onBackPressed()
        }
    }
}
