<script lang="ts">
    import { CheckCircle, AlertCircle, Mail } from "lucide-svelte";
    import Logo from "@components/Logo.svelte";

    interface Props {
        success?: string;
        error?: string;
    }

    let { success, error }: Props = $props();
</script>

<svelte:head>
    <title>Email Verification - Laju Go</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-neutral-50 dark:bg-neutral-950 px-4">
    <div class="max-w-md w-full">
        <div class="flex justify-center mb-8">
            <Logo />
        </div>

        <div class="bg-white dark:bg-neutral-900 rounded-2xl shadow-lg p-8 border border-neutral-200 dark:border-neutral-800">
            <div class="text-center">
                <div class="flex justify-center mb-4">
                    {#if success}
                        <div class="w-16 h-16 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center">
                            <CheckCircle class="w-8 h-8 text-green-600 dark:text-green-400" />
                        </div>
                    {:else if error}
                        <div class="w-16 h-16 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center">
                            <AlertCircle class="w-8 h-8 text-red-600 dark:text-red-400" />
                        </div>
                    {:else}
                        <div class="w-16 h-16 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center">
                            <Mail class="w-8 h-8 text-blue-600 dark:text-blue-400" />
                        </div>
                    {/if}
                </div>

                <h1 class="text-2xl font-bold text-neutral-900 dark:text-white mb-2">Email Verification</h1>

                {#if success}
                    <p class="text-green-600 dark:text-green-400 mb-6">{success}</p>
                    <a
                        href="/login"
                        class="inline-block w-full bg-brand-500 hover:bg-brand-600 text-white font-semibold py-3 px-6 rounded-lg transition-colors"
                    >
                        Continue to Login
                    </a>
                {:else if error}
                    <p class="text-red-600 dark:text-red-400 mb-6">{error}</p>
                    <a
                        href="/register"
                        class="inline-block w-full bg-neutral-200 hover:bg-neutral-300 dark:bg-neutral-800 dark:hover:bg-neutral-700 text-neutral-900 dark:text-white font-semibold py-3 px-6 rounded-lg transition-colors"
                    >
                        Back to Register
                    </a>
                {:else}
                    <p class="text-neutral-600 dark:text-neutral-400 mb-6">
                        Verifying your email...
                    </p>
                {/if}
            </div>
        </div>
    </div>
</div>
