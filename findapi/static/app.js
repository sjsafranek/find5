
function App(user, api) {
	this.api = api;
	this.views = {

		user: new Vue({
			el: '#user',
			delimiters: ['[|', '|]'],
			data: user,
			methods: {}
		}),

		sensor: new Vue({
			el: '#sensor',
			delimiters: ['[|', '|]'],
			data: {
				id: '',
				name: '',
				created_at: '',
				is_active: '',
				is_deleted: '',
				type: '',
				updated_at: ''
			},
			methods: {
				setData: function(data) {
					for (var i in data) {
						this.$set(this, i, data[i]);
					}
					$(this.$el).show();
				}
			}
		}),

		sensors: new Vue({
			el: '#sensors',
			delimiters: ['[|', '|]'],
			data: {
				device_id: '',
				sensors: []
			},
			methods: {
				setData: function(device_id, data) {
					var self = this;

					$(app.views.sensors.$el).hide();
					$(app.views.sensor.$el).hide();

					this.device_id = device_id;
					this.sensors = data.sensors = [];
					var $sensorsContainer = $(this.$el).find('.objects');
					$sensorsContainer.empty();
					data && data.map(function(d) {
						$sensorsContainer.append(
							$('<div>').append(
								$('<label>').append(d.name)
							).on('click', function(e){
								$(self.$el).find('.objects .selected').removeClass('selected');
								$(e.target).addClass('selected');
								app.views.sensor.setData(d);
							})
						);
					})
					$(this.$el).show();
				},
				createSensor: function() {
					var self = this;
					Swal.fire({
						title: 'Create Sensor',
						html:
							'<input id="swal-input1" class="swal2-input" placeholder="name">' +
							'<input id="swal-input2" class="swal2-input" placeholder="type">',
						focusConfirm: false,
						preConfirm: () => {
							return {
								'name': document.getElementById('swal-input1').value,
								'type': document.getElementById('swal-input2').value
							}
						}
					}).then((result) => {
						if (result.value) {
							api.createSensor(self.device_id, result.value.name, result.value.type, function(err, data) {
								if (err) return new swal('Error', JSON.stringify(err), 'error');
								api.getSensors(self.device_id, function(err, data) {
									if (err) return new swal('Error', JSON.stringify(err), 'error');
									self.setData(self.device_id, data.sensors || []);
								});
							});
						}
					});
				}
			}
		}),

		device: new Vue({
			el: '#device',
			delimiters: ['[|', '|]'],
			data: {
				id: '',
				name: '',
				created_at: '',
				is_active: '',
				is_deleted: '',
				type: '',
				updated_at: ''
			},
			methods: {
				setData: function(data) {
					for (var i in data) {
						this.$set(this, i, data[i]);
					}
					$(this.$el).show();
					app.views.sensors.setData(data.id, data.sensors);
				}
			}
		}),

		devices: new Vue({
			el: '#devices',
			delimiters: ['[|', '|]'],
			data: {devices: []},
			methods: {

				createDevice: function() {
					var self = this;
					Swal.fire({
						title: 'Create Device',
						html:
							'<input id="swal-input1" class="swal2-input" placeholder="name">' +
							'<input id="swal-input2" class="swal2-input" placeholder="type">',
						focusConfirm: false,
						preConfirm: () => {
							return {
								'name': document.getElementById('swal-input1').value,
								'type': document.getElementById('swal-input2').value
							}
						}
					}).then((result) => {
						if (result.value) {
							api.createDevice(result.value.name, result.value.type, function(err, data) {
								if (err) return new swal('Error', JSON.stringify(err), 'error');
								self.fetchDevices();
							});
						}
					});
				},

				fetchDevices: function() {
					var self = this;
					api.getDevices(function(err, data) {
						self.devices = data.devices || [];
						var $devicesContainer = $(self.$el).find('.objects');
						$devicesContainer.empty();
						self.devices.map(function(d) {
							$devicesContainer.append(
								$('<div>').append(
									$('<label>').append(d.name)
								).on('click', function(e){
									$(self.$el).find('.objects .selected').removeClass('selected');
									$(e.target).addClass('selected');
									app.views.device.setData(d);
								})
							);
						})
					});
				}

			}

		})
	}

	this.views.devices.fetchDevices();

}
