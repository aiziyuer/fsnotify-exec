require 'filewatcher'

Filewatcher.new(['README.rdoc']).watch do |changes|
    
    changes.each do |filename, event|
      puts "File #{event}: #{filename}"
    end
    
end

