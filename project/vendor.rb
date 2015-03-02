#!/usr/bin/env ruby

# This script is based on the BASH vendor script in the Docker source code.
# https://github.com/docker/docker/blob/fd2d45d7d465fe02f159f21389b92164dbb433d3/project/vendor.sh

require 'fileutils'

$root = File.join(File.dirname(__FILE__), '..')

def run(command)
	raise 'non-zero exit status' if !system(command)
end

def package(type, name, ref)

	path = File.join($root, 'vendor', 'src', name)	

	FileUtils.rm_rf path
	FileUtils.mkdir_p path

	case type

	when :git
		run "git clone --quiet --no-checkout https://#{name} #{path}"
		run "cd #{path} && git reset --quiet --hard #{ref}"

	when :hg
		run "hg clone --quiet --updaterev #{ref} https://#{name} #{path}"

	end

	run "rm -rf #{File.join(path, ".#{type}")}"
end

package :git, 'github.com/coreos/go-etcd', '6aa2da5a7a905609c93036b9307185a04a5a84a5'
