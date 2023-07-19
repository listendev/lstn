<table align=center>
  <tr>
    <td><b>critical</b> 🚨 3</td>
    <td><b>medium</b> ⚠️ 1</td>
    <td><b>low</b> 🔷 1</td>
  </tr>
</table>
🔍 The following behaviors have been detected in the dependency tree during installation.

<details>
<summary>
:stop_sign: <b>3</b> critical activities detected
</summary>

## <b><a href="https://verdicts.listen.dev/npm/foo/1.0.0">foo@1.0.0</a></b><br>





	

### :stop_sign: outbound network connection
<dl>
<dt>Dependency type</dt>
<dd>

Direct dependency

</dd>


<dt>Metadata</dt>
<dd>
<table>



<tr>
<td>commandline:</td><td>sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0</td>
</tr>



<tr>
<td>executable_path:</td><td>/bin/sh</td>
</tr>







<tr>
<td>parent_name:</td><td>node</td>
</tr>



	
</table>
</dd>
</dl>




## <b><a href="https://verdicts.listen.dev/npm/bar/1.0.0">bar@1.0.0</a></b><br>





	

### :stop_sign: outbound network connection
<dl>
<dt>Dependency type</dt>
<dd>



Transitive dependency  (<a href="https://verdicts.listen.dev/npm/foo/1.0.0">foo@1.0.0</a>)

</dd>


<dt>Metadata</dt>
<dd>
<table>



<tr>
<td>commandline:</td><td>sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0</td>
</tr>



<tr>
<td>executable_path:</td><td>/bin/sh</td>
</tr>







<tr>
<td>parent_name:</td><td>node</td>
</tr>



	
</table>
</dd>
</dl>




## <b><a href="https://verdicts.listen.dev/npm/foobar/1.0.0">foobar@1.0.0</a></b><br>





	

### :stop_sign: outbound network connection
<dl>
<dt>Dependency type</dt>
<dd>

Direct dependency

</dd>


<dt>Metadata</dt>
<dd>
<table>



<tr>
<td>commandline:</td><td>sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0</td>
</tr>



<tr>
<td>executable_path:</td><td>/bin/sh</td>
</tr>







<tr>
<td>parent_name:</td><td>node</td>
</tr>



	
</table>
</dd>
</dl>



</details>

<details>
<summary>
:warning: <b>1</b> medium activities detected
</summary>

## <b><a href="https://verdicts.listen.dev/npm/foobar/1.0.0">foobar@1.0.0</a></b><br>





	

### :warning: outbound network connection
<dl>
<dt>Dependency type</dt>
<dd>

Direct dependency

</dd>


<dt>Metadata</dt>
<dd>
<table>



<tr>
<td>commandline:</td><td>sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0</td>
</tr>



<tr>
<td>executable_path:</td><td>/bin/sh</td>
</tr>







<tr>
<td>parent_name:</td><td>node</td>
</tr>



	
</table>
</dd>
</dl>



</details>

<details>
<summary>
:large_blue_diamond: <b>1</b> low activities detected
</summary>


## <b><a href="https://verdicts.listen.dev/npm/foobar/1.0.0">foobar@1.0.0</a></b><br>





	

### :large_blue_diamond: outbound network connection
<dl>
<dt>Dependency type</dt>
<dd>

Direct dependency

</dd>


<dt>Metadata</dt>
<dd>
<table>



<tr>
<td>commandline:</td><td>sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0</td>
</tr>



<tr>
<td>executable_path:</td><td>/bin/sh</td>
</tr>







<tr>
<td>parent_name:</td><td>node</td>
</tr>



	
</table>
</dd>
</dl>



</details>

***
<i>Powered by</i> <b><a href="https://listen.dev">listen.dev</a> <img height=14 src="https://listen.dev/assets/images/dolphin-noborder.png"></b>