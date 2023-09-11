import React, { useState, useEffect } from 'react'

function User(params) {
    const [email,setEmail]=useState(""); 
	const [passw,setPassw]=useState(""); 
	const[dataInput, setDataInput]=useState(""); 
	const submitThis=()=>{
		const info={email:email,passw:passw}; 
		setDataInput([info]);
	}
	return(
    <section class="bg-white dark:bg-gray-900">
        <div class="container px-6 py-12 mx-auto">
            <div class="mt-8 lg:w-1/2 lg:mx-6">
                <div
                    class="w-full px-8 py-10 mx-auto overflow-hidden bg-white rounded-lg shadow-2xl dark:bg-gray-900 lg:max-w-xl shadow-gray-300/50 dark:shadow-black/50">
                        <h1 class="text-lg font-medium text-gray-700">What do you want to ask</h1>
                        <form class="mt-6"  action="" onSubmit={submitThis}>
                            <div class="flex-1">
                                <label class="block mb-2 text-sm text-gray-600 dark:text-gray-200">Email</label>
                                <input type="text" placeholder="John Doe"  action="" onSubmit={submitThis} class="block w-full px-5 py-3 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40" />
                            </div>

                            <div class="flex-1 mt-6">
                                <label class="block mb-2 text-sm text-gray-600 dark:text-gray-200">Password</label>
                                <input type="password" placeholder="***" name="passw" id="passw" value={passw} onChange={(e)=>setPassw(e.target.value)} class="block w-full px-5 py-3 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40" />
                            </div>
                            <button type="submit" class="w-full px-6 py-3 mt-6 text-sm font-medium tracking-wide text-white capitalize transition-colors duration-300 transform bg-blue-500 rounded-md hover:bg-blue-400 focus:outline-none focus:ring focus:ring-blue-300 focus:ring-opacity-50">
                            Login
                        </button>
                        </form>
                </div>
            </div>
        </div>
    </section>
    )
}

export default User;